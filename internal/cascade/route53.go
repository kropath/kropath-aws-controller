// Copyright 2026 kropath Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cascade

// Route53KropathSection holds the Route53-family governance fields from
// KropathConfig.spec.mandatory.route53 / .defaults.route53 (ADR-015 §3.5).
//
// Only 4 scalar fields are governed at the KropathConfig level for Route53.
// namingTemplate, syncedLabels, and syncedAnnotations are Route53Config-only.
//
// Zero value of each field is the permissive sentinel (not enforced).
type Route53KropathSection struct {
	// DefaultTTL is the org-wide default TTL in seconds for DNS records.
	// 0 = not enforced; first-non-zero-wins in cascade. Does not apply to alias records.
	DefaultTTL int64 `json:"defaultTTL,omitempty"`

	// HealthCheckRequestInterval is the health check polling interval in seconds.
	// 0 = not enforced. Valid non-zero values: 10 (fast) or 30 (standard).
	HealthCheckRequestInterval int64 `json:"healthCheckRequestInterval,omitempty"`

	// HealthCheckFailureThreshold is the number of consecutive failures before a
	// health check is marked unhealthy. 0 = not enforced. Valid non-zero range: 1–10.
	HealthCheckFailureThreshold int64 `json:"healthCheckFailureThreshold,omitempty"`

	// ResolverEndpointType constrains the IP address family for Resolver endpoints.
	// Empty string = not enforced. Valid: "IPV4", "IPV6", "DUALSTACK".
	ResolverEndpointType string `json:"resolverEndpointType,omitempty"`

	// Tags are tier-level cloud resource tags.
	// The reconciler populates this from KropathConfig.spec.mandatory.tags / .defaults.tags
	// so that tag union merge flows through MergeRoute53Cascade alongside the
	// Route53-specific fields.
	// nil / empty map = no tags at this level.
	Tags map[string]string `json:"tags,omitempty"`
}

// Route53ConfigSection holds the Route53 governance fields from
// Route53Config.spec.mandatory or Route53Config.spec.defaults
// (per-type ResourceConfig, ADR-015 §3.5).
//
// Zero value of each field is the permissive sentinel (not enforced).
type Route53ConfigSection struct {
	// DefaultTTL is the default TTL in seconds for DNS records. 0 = not enforced.
	DefaultTTL int64 `json:"defaultTTL,omitempty"`

	// HealthCheckRequestInterval is the health check polling interval in seconds.
	// 0 = not enforced. Valid non-zero values: 10 or 30.
	HealthCheckRequestInterval int64 `json:"healthCheckRequestInterval,omitempty"`

	// HealthCheckFailureThreshold is the consecutive-failure count for unhealthy status.
	// 0 = not enforced. Valid non-zero range: 1–10.
	HealthCheckFailureThreshold int64 `json:"healthCheckFailureThreshold,omitempty"`

	// ResolverEndpointType constrains the IP address family for Resolver endpoints.
	// Empty string = not enforced. Valid: "IPV4", "IPV6", "DUALSTACK".
	ResolverEndpointType string `json:"resolverEndpointType,omitempty"`

	// NamingTemplate is the resource naming template (e.g. "{namespace}-{name}").
	// Governed only at Route53Config levels 3-4 (mandatory) and 6-7 (defaults).
	// KropathConfig.route53 does NOT carry namingTemplate.
	// Empty string = not enforced.
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// SyncedLabels are Kubernetes labels to propagate to created Route53 resources.
	// Additive map merge across Route53Config tiers only.
	SyncedLabels map[string]string `json:"syncedLabels,omitempty"`

	// SyncedAnnotations are Kubernetes annotations to propagate to created Route53 resources.
	// Additive map merge across Route53Config tiers only.
	SyncedAnnotations map[string]string `json:"syncedAnnotations,omitempty"`

	// Tags are cloud resource tags for this Route53 config profile.
	Tags map[string]string `json:"tags,omitempty"`
}

// EffectiveRoute53Section is one tier (mandatory or defaults) of the merged
// Route53 governance result written into Route53Config.status.effectiveConfig
// by the controller.
type EffectiveRoute53Section struct {
	DefaultTTL                  int64             `json:"defaultTTL,omitempty"`
	HealthCheckRequestInterval  int64             `json:"healthCheckRequestInterval,omitempty"`
	HealthCheckFailureThreshold int64             `json:"healthCheckFailureThreshold,omitempty"`
	ResolverEndpointType        string            `json:"resolverEndpointType,omitempty"`
	NamingTemplate              string            `json:"namingTemplate,omitempty"`
	SyncedLabels                map[string]string `json:"syncedLabels,omitempty"`
	SyncedAnnotations           map[string]string `json:"syncedAnnotations,omitempty"`
	Tags                        map[string]string `json:"tags,omitempty"`
}

// EffectiveRoute53Config is the merged Route53 governance result written into
// Route53Config.status.effectiveConfig by the controller.
type EffectiveRoute53Config struct {
	Mandatory EffectiveRoute53Section `json:"mandatory"`
	Defaults  EffectiveRoute53Section `json:"defaults"`
}

// MergeRoute53Cascade merges Route53 governance fields from all cascade
// sources and returns the effective configuration to be written to status.effectiveConfig.
//
// Ten-level priority chain for Route53 (ADR-015 §5.3):
//
//	Level 1 — globalKropathMandatory  (KropathConfig in kro-system, mandatory.route53)
//	Level 2 — localKropathMandatory   (KropathConfig in resource namespace, mandatory.route53)
//	Level 3 — globalR53CfgMandatory   (Route53Config in kro-system, mandatory)
//	Level 4 — localR53CfgMandatory    (Route53Config in resource namespace, mandatory)
//	(Level 5 = instance spec — resolved in RGD CEL, not here)
//	Level 6 — localR53CfgDefaults     (Route53Config in resource namespace, defaults)
//	Level 7 — globalR53CfgDefaults    (Route53Config in kro-system, defaults)
//	Level 8 — localKropathDefaults    (KropathConfig in resource namespace, defaults.route53)
//	Level 9 — globalKropathDefaults   (KropathConfig in kro-system, defaults.route53)
//
// Scalar merge: firstNonEmptyString / firstNonZeroInt64 in priority order (lowest number wins).
// Tags: additive union merge across all four mandatory levels, all four defaults levels.
// SyncedLabels/SyncedAnnotations: additive union from Route53Config levels only (no KropathConfig).
func MergeRoute53Cascade(
	globalKropathMandatory Route53KropathSection, // level 1
	localKropathMandatory Route53KropathSection, // level 2
	globalR53CfgMandatory Route53ConfigSection, // level 3
	localR53CfgMandatory Route53ConfigSection, // level 4
	localR53CfgDefaults Route53ConfigSection, // level 6
	globalR53CfgDefaults Route53ConfigSection, // level 7
	localKropathDefaults Route53KropathSection, // level 8
	globalKropathDefaults Route53KropathSection, // level 9
) EffectiveRoute53Config {
	return EffectiveRoute53Config{
		Mandatory: EffectiveRoute53Section{
			DefaultTTL: firstNonZeroInt64(
				globalKropathMandatory.DefaultTTL,
				localKropathMandatory.DefaultTTL,
				globalR53CfgMandatory.DefaultTTL,
				localR53CfgMandatory.DefaultTTL,
			),
			HealthCheckRequestInterval: firstNonZeroInt64(
				globalKropathMandatory.HealthCheckRequestInterval,
				localKropathMandatory.HealthCheckRequestInterval,
				globalR53CfgMandatory.HealthCheckRequestInterval,
				localR53CfgMandatory.HealthCheckRequestInterval,
			),
			HealthCheckFailureThreshold: firstNonZeroInt64(
				globalKropathMandatory.HealthCheckFailureThreshold,
				localKropathMandatory.HealthCheckFailureThreshold,
				globalR53CfgMandatory.HealthCheckFailureThreshold,
				localR53CfgMandatory.HealthCheckFailureThreshold,
			),
			ResolverEndpointType: firstNonEmptyString(
				globalKropathMandatory.ResolverEndpointType,
				localKropathMandatory.ResolverEndpointType,
				globalR53CfgMandatory.ResolverEndpointType,
				localR53CfgMandatory.ResolverEndpointType,
			),
			// NamingTemplate: Route53Config levels only (3, 4); KropathConfig has no namingTemplate.
			NamingTemplate: firstNonEmptyString(
				globalR53CfgMandatory.NamingTemplate,
				localR53CfgMandatory.NamingTemplate,
			),
			// SyncedLabels: additive union from Route53Config levels only.
			// L4 added first (lowest priority), L3 wins on key conflict.
			SyncedLabels: mergeMaps(
				localR53CfgMandatory.SyncedLabels,
				globalR53CfgMandatory.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				localR53CfgMandatory.SyncedAnnotations,
				globalR53CfgMandatory.SyncedAnnotations,
			),
			// Tags: union of all mandatory sources; L4 added first, L1 wins on key conflict.
			Tags: mergeMaps(
				localR53CfgMandatory.Tags,
				globalR53CfgMandatory.Tags,
				localKropathMandatory.Tags,
				globalKropathMandatory.Tags,
			),
		},
		Defaults: EffectiveRoute53Section{
			DefaultTTL: firstNonZeroInt64(
				localR53CfgDefaults.DefaultTTL,
				globalR53CfgDefaults.DefaultTTL,
				localKropathDefaults.DefaultTTL,
				globalKropathDefaults.DefaultTTL,
			),
			HealthCheckRequestInterval: firstNonZeroInt64(
				localR53CfgDefaults.HealthCheckRequestInterval,
				globalR53CfgDefaults.HealthCheckRequestInterval,
				localKropathDefaults.HealthCheckRequestInterval,
				globalKropathDefaults.HealthCheckRequestInterval,
			),
			HealthCheckFailureThreshold: firstNonZeroInt64(
				localR53CfgDefaults.HealthCheckFailureThreshold,
				globalR53CfgDefaults.HealthCheckFailureThreshold,
				localKropathDefaults.HealthCheckFailureThreshold,
				globalKropathDefaults.HealthCheckFailureThreshold,
			),
			ResolverEndpointType: firstNonEmptyString(
				localR53CfgDefaults.ResolverEndpointType,
				globalR53CfgDefaults.ResolverEndpointType,
				localKropathDefaults.ResolverEndpointType,
				globalKropathDefaults.ResolverEndpointType,
			),
			// NamingTemplate: Route53Config levels only (6, 7).
			NamingTemplate: firstNonEmptyString(
				localR53CfgDefaults.NamingTemplate,
				globalR53CfgDefaults.NamingTemplate,
			),
			// SyncedLabels: additive union from Route53Config levels only.
			// L7 added first (lowest priority), L6 wins on key conflict.
			SyncedLabels: mergeMaps(
				globalR53CfgDefaults.SyncedLabels,
				localR53CfgDefaults.SyncedLabels,
			),
			// SyncedAnnotations: same additive pattern as SyncedLabels.
			SyncedAnnotations: mergeMaps(
				globalR53CfgDefaults.SyncedAnnotations,
				localR53CfgDefaults.SyncedAnnotations,
			),
			// Tags: union of all defaults sources; L9 added first, L6 wins on key conflict.
			Tags: mergeMaps(
				globalKropathDefaults.Tags,
				localKropathDefaults.Tags,
				globalR53CfgDefaults.Tags,
				localR53CfgDefaults.Tags,
			),
		},
	}
}

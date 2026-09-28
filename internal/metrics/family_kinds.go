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

package metrics

import (
	"context"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// familyConfigKind names one of the 57 <Family>Config kinds and how to build
// an empty list object for it. This table is a literal rather than a scheme
// lookup (as internal/reconciler/kropathconfigstatus uses) because
// internal/metrics must not import internal/features or the reconciler
// packages that import internal/metrics -- doing so from here would cycle.
// TestFamilyConfigKindsMatchesFeatures keeps it in sync with features.All.
type familyConfigKind struct {
	kind    string
	family  string
	newList func() client.ObjectList
}

var familyConfigKinds = []familyConfigKind{
	{kind: "IAMConfig", family: "iamconfig", newList: func() client.ObjectList { return &v1alpha1.IAMConfigList{} }},
	{kind: "S3Config", family: "s3config", newList: func() client.ObjectList { return &v1alpha1.S3ConfigList{} }},
	{kind: "KMSConfig", family: "kmsconfig", newList: func() client.ObjectList { return &v1alpha1.KMSConfigList{} }},
	{kind: "SQSConfig", family: "sqsconfig", newList: func() client.ObjectList { return &v1alpha1.SQSConfigList{} }},
	{kind: "SecretsManagerConfig", family: "secretsmanagerconfig", newList: func() client.ObjectList { return &v1alpha1.SecretsManagerConfigList{} }},
	{kind: "SNSConfig", family: "snsconfig", newList: func() client.ObjectList { return &v1alpha1.SNSConfigList{} }},
	{kind: "DynamoDBConfig", family: "dynamodbconfig", newList: func() client.ObjectList { return &v1alpha1.DynamoDBConfigList{} }},
	{kind: "EventBridgeConfig", family: "eventbridgeconfig", newList: func() client.ObjectList { return &v1alpha1.EventBridgeConfigList{} }},
	{kind: "CloudWatchLogsConfig", family: "cloudwatchlogsconfig", newList: func() client.ObjectList { return &v1alpha1.CloudWatchLogsConfigList{} }},
	{kind: "CloudWatchConfig", family: "cloudwatchconfig", newList: func() client.ObjectList { return &v1alpha1.CloudWatchConfigList{} }},
	{kind: "ELBConfig", family: "elbconfig", newList: func() client.ObjectList { return &v1alpha1.ELBConfigList{} }},
	{kind: "RDSConfig", family: "rdsconfig", newList: func() client.ObjectList { return &v1alpha1.RDSConfigList{} }},
	{kind: "AutoScalingConfig", family: "autoscalingconfig", newList: func() client.ObjectList { return &v1alpha1.AutoScalingConfigList{} }},
	{kind: "ECSConfig", family: "ecsconfig", newList: func() client.ObjectList { return &v1alpha1.ECSConfigList{} }},
	{kind: "EKSConfig", family: "eksconfig", newList: func() client.ObjectList { return &v1alpha1.EKSConfigList{} }},
	{kind: "EC2Config", family: "ec2config", newList: func() client.ObjectList { return &v1alpha1.EC2ConfigList{} }},
	{kind: "ApiGatewayV2Config", family: "apigatewayv2config", newList: func() client.ObjectList { return &v1alpha1.ApiGatewayV2ConfigList{} }},
	{kind: "APIGatewayConfig", family: "apigatewayconfig", newList: func() client.ObjectList { return &v1alpha1.APIGatewayConfigList{} }},
	{kind: "EFSConfig", family: "efsconfig", newList: func() client.ObjectList { return &v1alpha1.EFSConfigList{} }},
	{kind: "ElastiCacheConfig", family: "elasticacheconfig", newList: func() client.ObjectList { return &v1alpha1.ElastiCacheConfigList{} }},
	{kind: "ECRConfig", family: "ecrconfig", newList: func() client.ObjectList { return &v1alpha1.ECRConfigList{} }},
	{kind: "StepFunctionsConfig", family: "stepfunctionsconfig", newList: func() client.ObjectList { return &v1alpha1.StepFunctionsConfigList{} }},
	{kind: "MSKConfig", family: "mskconfig", newList: func() client.ObjectList { return &v1alpha1.MSKConfigList{} }},
	{kind: "MemoryDBConfig", family: "memorydbconfig", newList: func() client.ObjectList { return &v1alpha1.MemoryDBConfigList{} }},
	{kind: "ACMConfig", family: "acmconfig", newList: func() client.ObjectList { return &v1alpha1.ACMConfigList{} }},
	{kind: "EMRConfig", family: "emrconfig", newList: func() client.ObjectList { return &v1alpha1.EMRConfigList{} }},
	{kind: "DocumentDBConfig", family: "documentdbconfig", newList: func() client.ObjectList { return &v1alpha1.DocumentDBConfigList{} }},
	{kind: "GlueConfig", family: "glueconfig", newList: func() client.ObjectList { return &v1alpha1.GlueConfigList{} }},
	{kind: "AthenaConfig", family: "athenaconfig", newList: func() client.ObjectList { return &v1alpha1.AthenaConfigList{} }},
	{kind: "DSQLConfig", family: "dsqlconfig", newList: func() client.ObjectList { return &v1alpha1.DSQLConfigList{} }},
	{kind: "Route53Config", family: "route53config", newList: func() client.ObjectList { return &v1alpha1.Route53ConfigList{} }},
	{kind: "SSMConfig", family: "ssmconfig", newList: func() client.ObjectList { return &v1alpha1.SSMConfigList{} }},
	{kind: "CognitoConfig", family: "cognitoconfig", newList: func() client.ObjectList { return &v1alpha1.CognitoConfigList{} }},
	{kind: "KinesisConfig", family: "kinesisconfig", newList: func() client.ObjectList { return &v1alpha1.KinesisConfigList{} }},
	{kind: "CloudTrailConfig", family: "cloudtrailconfig", newList: func() client.ObjectList { return &v1alpha1.CloudTrailConfigList{} }},
	{kind: "AppScalingConfig", family: "appscalingconfig", newList: func() client.ObjectList { return &v1alpha1.AppScalingConfigList{} }},
	{kind: "KeyspacesConfig", family: "keyspacesconfig", newList: func() client.ObjectList { return &v1alpha1.KeyspacesConfigList{} }},
	{kind: "WAFConfig", family: "wafconfig", newList: func() client.ObjectList { return &v1alpha1.WAFConfigList{} }},
	{kind: "BedrockConfig", family: "bedrockconfig", newList: func() client.ObjectList { return &v1alpha1.BedrockConfigList{} }},
	{kind: "SageMakerConfig", family: "sagemakerconfig", newList: func() client.ObjectList { return &v1alpha1.SageMakerConfigList{} }},
	{kind: "OpenSearchConfig", family: "opensearchconfig", newList: func() client.ObjectList { return &v1alpha1.OpenSearchConfigList{} }},
	{kind: "PipesConfig", family: "pipesconfig", newList: func() client.ObjectList { return &v1alpha1.PipesConfigList{} }},
	{kind: "SESConfig", family: "sesconfig", newList: func() client.ObjectList { return &v1alpha1.SESConfigList{} }},
	{kind: "CodeArtifactConfig", family: "codeartifactconfig", newList: func() client.ObjectList { return &v1alpha1.CodeArtifactConfigList{} }},
	{kind: "MWAAConfig", family: "mwaaconfig", newList: func() client.ObjectList { return &v1alpha1.MWAAConfigList{} }},
	{kind: "NetworkFirewallConfig", family: "networkfirewallconfig", newList: func() client.ObjectList { return &v1alpha1.NetworkFirewallConfigList{} }},
	{kind: "BackupConfig", family: "backupconfig", newList: func() client.ObjectList { return &v1alpha1.BackupConfigList{} }},
	{kind: "OrganizationsConfig", family: "organizationsconfig", newList: func() client.ObjectList { return &v1alpha1.OrganizationsConfigList{} }},
	{kind: "ManagedPrometheusConfig", family: "managedprometheusconfig", newList: func() client.ObjectList { return &v1alpha1.ManagedPrometheusConfigList{} }},
	{kind: "RAMConfig", family: "ramconfig", newList: func() client.ObjectList { return &v1alpha1.RAMConfigList{} }},
	{kind: "MQConfig", family: "mqconfig", newList: func() client.ObjectList { return &v1alpha1.MQConfigList{} }},
	{kind: "QuickSightConfig", family: "quicksightconfig", newList: func() client.ObjectList { return &v1alpha1.QuickSightConfigList{} }},
	{kind: "ECRPublicConfig", family: "ecrpublicconfig", newList: func() client.ObjectList { return &v1alpha1.ECRPublicConfigList{} }},
	{kind: "RecycleBinConfig", family: "recyclebinconfig", newList: func() client.ObjectList { return &v1alpha1.RecycleBinConfigList{} }},
	{kind: "S3AdvancedConfig", family: "s3advancedconfig", newList: func() client.ObjectList { return &v1alpha1.S3AdvancedConfigList{} }},
	{kind: "CloudFrontConfig", family: "cloudfrontconfig", newList: func() client.ObjectList { return &v1alpha1.CloudFrontConfigList{} }},
	{kind: "LambdaConfig", family: "lambdaconfig", newList: func() client.ObjectList { return &v1alpha1.LambdaConfigList{} }},
}

// countUnavailableFamilyKinds lists every <Family>Config kind via apiReader
// and counts the ones whose CRD is absent (NoKindMatchError) -- "not
// installed", not a collect failure (spec §3, §3.4). A real List error for any
// other kind aborts the count and is returned to the caller, which must
// attribute it to its own collector label.
func countUnavailableFamilyKinds(ctx context.Context, apiReader client.Reader) (int, error) {
	unavailable := 0
	for _, fk := range familyConfigKinds {
		if err := apiReader.List(ctx, fk.newList()); err != nil {
			if apimeta.IsNoMatchError(err) {
				unavailable++
				continue
			}
			return 0, err
		}
	}
	return unavailable, nil
}

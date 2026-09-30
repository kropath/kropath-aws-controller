// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/acmconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/apigatewayconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/apigatewayv2config"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/appscalingconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/athenaconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/autoscalingconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/backupconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/bedrockconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/cloudfrontconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/cloudtrailconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/cloudwatchconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/cloudwatchlogsconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/codeartifactconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/cognitoconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/documentdbconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/dsqlconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/dynamodbconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/ec2config"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/ecrconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/ecrpublicconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/ecsconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/efsconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/eksconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/elasticacheconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/elbconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/emrconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/eventbridgeconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/glueconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/iamconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/keyspacesconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/kinesisconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/kmsconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/kropathconfigstatus"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/labeloperator"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/lambdaconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/managedprometheusconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/memorydbconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/mqconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/mskconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/mwaaconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/namespaceplacement"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/networkfirewallconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/opensearchconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/organizationsconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/pipesconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/policydocument"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/quicksightconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/ramconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/rdsconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/recyclebinconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/route53config"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/s3advancedconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/s3config"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/sagemakerconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/secretsmanagerconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/sesconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/snsconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/sqsconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/ssmconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/stepfunctionsconfig"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/wafconfig"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const awsGroup = "aws.kropath.run"
const awsVersion = "v1alpha1"

func awsGVK(kind string) schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: awsGroup, Version: awsVersion, Kind: kind}
}

// policyDocumentRefGVKs are the six optional GVKs that policydocument watches for
// ref resolution. Commit 2 (gate) will gate some of these on CRD availability.
var policyDocumentRefGVKs = []schema.GroupVersionKind{
	awsGVK("AWSIAMRole"),
	awsGVK("AWSS3Bucket"),
	awsGVK("AWSLambdaFunction"),
	awsGVK("AWSSQSQueue"),
	awsGVK("AWSKMSKey"),
	awsGVK("AWSSecretsManagerSecret"),
}

// All returns the complete registry in the same order as features.All.
func All() []Entry {
	// pdr is declared here so both the Build and AddKindWatch closures for
	// policydocument capture the same pointer variable. Build sets it on first
	// call; AddKindWatch reads it thereafter. The coordinator mutex guarantees
	// that Build completes before any AddKindWatch call is dispatched.
	var pdr *policydocument.Reconciler

	// labelop* vars are captured by both the labeloperator Build and AddKindWatch
	// closures. Build populates all three; AddKindWatch reads them.
	var (
		labelopHandles map[string]controller.Controller
		labelopMgr     ctrl.Manager
		labelopLog     logr.Logger
	)

	return []Entry{
		cascadeEntry("iamconfig", "IAMConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&iamconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("IAMConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("s3config", "S3Config", func(bctx BuildCtx) (controller.Controller, error) {
			return (&s3config.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("S3Config"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("kmsconfig", "KMSConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&kmsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("KMSConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("sqsconfig", "SQSConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&sqsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("SQSConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("secretsmanagerconfig", "SecretsManagerConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&secretsmanagerconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("SecretsManagerConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		// PolicyDocument watches its own kind plus six optional ref GVKs.
		// The closure captures pdr declared above so AddKindWatch can find the
		// already-built reconciler without a separate lookup.
		{
			Package:  "policydocument",
			Required: []schema.GroupVersionKind{awsGVK("PolicyDocument"), awsGVK("KropathConfig")},
			Optional: policyDocumentRefGVKs,
			Build: func(bctx BuildCtx, servedOptional []schema.GroupVersionKind) (controller.Controller, error) {
				pdr = &policydocument.Reconciler{}
				return pdr.BuildWithManager(bctx.Manager, servedOptional)
			},
			AddKindWatch: func(c controller.Controller, gvk schema.GroupVersionKind) error {
				if pdr == nil {
					return fmt.Errorf("policydocument: AddKindWatch called before Build")
				}
				return pdr.AddKindWatch(c, gvk)
			},
		},
		// LabelOperator creates one controller per provider GVK at startup, and registers
		// new controllers at runtime via AddKindWatch when new provider CRDs appear.
		// Optional == nil signals wildcard: any provider GVK triggers AddKindWatch.
		{
			Package:  "labeloperator",
			Required: nil,
			Optional: nil, // wildcard: handled by registry.OnGVKServable
			Build: func(bctx BuildCtx, _ []schema.GroupVersionKind) (controller.Controller, error) {
				log := bctx.Log.WithName("controllers").WithName("LabelOperator")
				handles, err := labeloperator.Setup(bctx.Manager, log)
				if err != nil {
					return nil, fmt.Errorf("label-operator setup: %w", err)
				}
				labelopHandles = handles
				labelopMgr = bctx.Manager
				labelopLog = log
				return nil, nil
			},
			AddKindWatch: func(_ controller.Controller, gvk schema.GroupVersionKind) error {
				if labelopHandles == nil {
					return fmt.Errorf("labeloperator: AddKindWatch called before Build")
				}
				if labeloperator.LabelKeyForGroup(gvk.Group) == "" {
					return nil // not a watched provider group
				}
				key := gvk.String()
				if labelopHandles[key] != nil {
					return nil // controller already registered at startup
				}
				handle, err := labeloperator.RegisterController(
					labelopMgr, labeloperator.ControllerName(gvk), gvk, labelopLog,
				)
				if err != nil {
					return fmt.Errorf("label-operator: register runtime controller for %v: %w", gvk, err)
				}
				labelopHandles[key] = handle
				return nil
			},
		},
		cascadeEntry("snsconfig", "SNSConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&snsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("SNSConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("dynamodbconfig", "DynamoDBConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&dynamodbconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("DynamoDBConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("eventbridgeconfig", "EventBridgeConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&eventbridgeconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("EventBridgeConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("cloudwatchlogsconfig", "CloudWatchLogsConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&cloudwatchlogsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("CloudWatchLogsConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("cognitoconfig", "CognitoConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&cognitoconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("CognitoConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("cloudwatchconfig", "CloudWatchConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&cloudwatchconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("CloudWatchConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("elbconfig", "ELBConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&elbconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("AWSELBConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("rdsconfig", "RDSConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&rdsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("RDSConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("autoscalingconfig", "AutoScalingConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&autoscalingconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("AutoScalingConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("ecsconfig", "ECSConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&ecsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("ECSConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("eksconfig", "EKSConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&eksconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("EKSConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("ec2config", "EC2Config", func(bctx BuildCtx) (controller.Controller, error) {
			return (&ec2config.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("EC2Config"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("apigatewayv2config", "ApiGatewayV2Config", func(bctx BuildCtx) (controller.Controller, error) {
			return (&apigatewayv2config.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("ApiGatewayV2Config"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("apigatewayconfig", "APIGatewayConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&apigatewayconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("APIGatewayConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("efsconfig", "EFSConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&efsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("EFSConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("elasticacheconfig", "ElastiCacheConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&elasticacheconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("ElastiCacheConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("ecrconfig", "ECRConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&ecrconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("ECRConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("stepfunctionsconfig", "StepFunctionsConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&stepfunctionsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("StepFunctionsConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("mskconfig", "MSKConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&mskconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("MSKConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("memorydbconfig", "MemoryDBConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&memorydbconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("MemoryDBConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("acmconfig", "ACMConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&acmconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("ACMConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("emrconfig", "EMRConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&emrconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("EMRConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("glueconfig", "GlueConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&glueconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("GlueConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("documentdbconfig", "DocumentDBConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&documentdbconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("DocumentDBConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("athenaconfig", "AthenaConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&athenaconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("AthenaConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("dsqlconfig", "DSQLConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&dsqlconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("DSQLConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("route53config", "Route53Config", func(bctx BuildCtx) (controller.Controller, error) {
			return (&route53config.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("Route53Config"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("ssmconfig", "SSMConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&ssmconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("SSMConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("kinesisconfig", "KinesisConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&kinesisconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("KinesisConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("cloudtrailconfig", "CloudTrailConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&cloudtrailconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("CloudTrailConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("appscalingconfig", "AppScalingConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&appscalingconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("AppScalingConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("keyspacesconfig", "KeyspacesConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&keyspacesconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("KeyspacesConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("wafconfig", "WAFConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&wafconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("WAFConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("bedrockconfig", "BedrockConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&bedrockconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("BedrockConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("sagemakerconfig", "SageMakerConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&sagemakerconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("SageMakerConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("opensearchconfig", "OpenSearchConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&opensearchconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("OpenSearchConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("pipesconfig", "PipesConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&pipesconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("PipesConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("sesconfig", "SESConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&sesconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("SESConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("codeartifactconfig", "CodeArtifactConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&codeartifactconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("CodeArtifactConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("mwaaconfig", "MWAAConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&mwaaconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("MWAAConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("networkfirewallconfig", "NetworkFirewallConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&networkfirewallconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("NetworkFirewallConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("backupconfig", "BackupConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&backupconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("BackupConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("organizationsconfig", "OrganizationsConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&organizationsconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("OrganizationsConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("managedprometheusconfig", "ManagedPrometheusConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&managedprometheusconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("ManagedPrometheusConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("ramconfig", "RAMConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&ramconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("RAMConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("mqconfig", "MQConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&mqconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("MQConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("quicksightconfig", "QuickSightConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&quicksightconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("QuickSightConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("ecrpublicconfig", "ECRPublicConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&ecrpublicconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("ECRPublicConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("recyclebinconfig", "RecycleBinConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&recyclebinconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("RecycleBinConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("s3advancedconfig", "S3AdvancedConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&s3advancedconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("S3AdvancedConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("cloudfrontconfig", "CloudFrontConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&cloudfrontconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("CloudFrontConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		cascadeEntry("lambdaconfig", "LambdaConfig", func(bctx BuildCtx) (controller.Controller, error) {
			return (&lambdaconfig.Reconciler{
				Client: bctx.Manager.GetClient(),
				Log:    bctx.Log.WithName("controllers").WithName("LambdaConfig"),
				Scheme: bctx.Manager.GetScheme(),
			}).BuildWithManager(bctx.Manager)
		}),
		// KropathConfigStatus watches only KropathConfig itself — it derives the
		// family kinds it lists from features.All at reconcile time rather than
		// declaring 57 Required GVKs, so a family whose CRD is briefly absent
		// degrades that one kind's consumer count instead of staying pending.
		{
			Package:  "kropathconfigstatus",
			Required: []schema.GroupVersionKind{awsGVK("KropathConfig")},
			Optional: nil,
			Build: func(bctx BuildCtx, _ []schema.GroupVersionKind) (controller.Controller, error) {
				return (&kropathconfigstatus.Reconciler{
					Client: bctx.Manager.GetClient(),
					Reader: bctx.Manager.GetAPIReader(),
					Log:    bctx.Log.WithName("controllers").WithName("KropathConfigStatus"),
					Scheme: bctx.Manager.GetScheme(),
				}).BuildWithManager(bctx.Manager)
			},
		},
		// NamespacePlacement watches only the core Namespace kind (never gated on
		// any aws.kropath.run CRD) and publishes the account/region placement
		// verdict as an annotation + Event (spec §5.6, §6.2).
		{
			Package:  "namespaceplacement",
			Required: nil,
			Optional: nil,
			Build: func(bctx BuildCtx, _ []schema.GroupVersionKind) (controller.Controller, error) {
				return (&namespaceplacement.Reconciler{
					Client: bctx.Manager.GetClient(),
					//nolint:staticcheck // GetEventRecorder's newer Eventf API requires an "action" field the spec doesn't define; GetEventRecorderFor is still fully supported.
					Recorder: bctx.Manager.GetEventRecorderFor("kropath-placement"),
					Log:      bctx.Log.WithName("controllers").WithName("NamespacePlacement"),
					Scheme:   bctx.Manager.GetScheme(),
				}).BuildWithManager(bctx.Manager)
			},
		},
	}
}

// cascadeEntry builds a standard cascade reconciler entry. Every cascade reconciler
// watches its own <Family>Config plus KropathConfig. The inner build function does
// not receive servedOptional because cascade reconcilers have no Optional GVKs.
func cascadeEntry(pkg, kind string, build func(BuildCtx) (controller.Controller, error)) Entry {
	return Entry{
		Package:  pkg,
		Required: []schema.GroupVersionKind{awsGVK(kind), awsGVK("KropathConfig")},
		Optional: nil,
		Build: func(bctx BuildCtx, _ []schema.GroupVersionKind) (controller.Controller, error) {
			return build(bctx)
		},
		AddKindWatch: nil,
	}
}

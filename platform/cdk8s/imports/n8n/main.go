// n8n
package n8n

import (
	"reflect"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func init() {
	_jsii_.RegisterStruct(
		"n8n.N8NApi",
		reflect.TypeOf((*N8NApi)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NApiSwagger",
		reflect.TypeOf((*N8NApiSwagger)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NBinaryData",
		reflect.TypeOf((*N8NBinaryData)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NBinaryDataAvailableModes",
		reflect.TypeOf((*N8NBinaryDataAvailableModes)(nil)).Elem(),
		map[string]interface{}{
			"FILESYSTEM": N8NBinaryDataAvailableModes_FILESYSTEM,
			"S3": N8NBinaryDataAvailableModes_S3,
		},
	)
	_jsii_.RegisterEnum(
		"n8n.N8NBinaryDataMode",
		reflect.TypeOf((*N8NBinaryDataMode)(nil)).Elem(),
		map[string]interface{}{
			"DEFAULT": N8NBinaryDataMode_DEFAULT,
			"FILESYSTEM": N8NBinaryDataMode_FILESYSTEM,
			"S3": N8NBinaryDataMode_S3,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NBinaryDataS3",
		reflect.TypeOf((*N8NBinaryDataS3)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NDb",
		reflect.TypeOf((*N8NDb)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NDbLogging",
		reflect.TypeOf((*N8NDbLogging)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NDbLoggingOptions",
		reflect.TypeOf((*N8NDbLoggingOptions)(nil)).Elem(),
		map[string]interface{}{
			"QUERY": N8NDbLoggingOptions_QUERY,
			"ERROR": N8NDbLoggingOptions_ERROR,
			"SCHEMA": N8NDbLoggingOptions_SCHEMA,
			"WARN": N8NDbLoggingOptions_WARN,
			"INFO": N8NDbLoggingOptions_INFO,
			"LOG": N8NDbLoggingOptions_LOG,
			"ALL": N8NDbLoggingOptions_ALL,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NDbPostgresdb",
		reflect.TypeOf((*N8NDbPostgresdb)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NDbPostgresdbSsl",
		reflect.TypeOf((*N8NDbPostgresdbSsl)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NDbPostgresdbSslExistingCertFileSecret",
		reflect.TypeOf((*N8NDbPostgresdbSslExistingCertFileSecret)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NDbPostgresdbSslExistingCertificateAuthorityFileSecret",
		reflect.TypeOf((*N8NDbPostgresdbSslExistingCertificateAuthorityFileSecret)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NDbPostgresdbSslExistingPrivateKeyFileSecret",
		reflect.TypeOf((*N8NDbPostgresdbSslExistingPrivateKeyFileSecret)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NDbSqlite",
		reflect.TypeOf((*N8NDbSqlite)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NDbType",
		reflect.TypeOf((*N8NDbType)(nil)).Elem(),
		map[string]interface{}{
			"SQLITE": N8NDbType_SQLITE,
			"POSTGRESDB": N8NDbType_POSTGRESDB,
		},
	)
	_jsii_.RegisterEnum(
		"n8n.N8NDefaultLocale",
		reflect.TypeOf((*N8NDefaultLocale)(nil)).Elem(),
		map[string]interface{}{
			"AF": N8NDefaultLocale_AF,
			"AM": N8NDefaultLocale_AM,
			"AS": N8NDefaultLocale_AS,
			"BE": N8NDefaultLocale_BE,
			"BG": N8NDefaultLocale_BG,
			"BS": N8NDefaultLocale_BS,
			"CA": N8NDefaultLocale_CA,
			"CS": N8NDefaultLocale_CS,
			"CY": N8NDefaultLocale_CY,
			"DA": N8NDefaultLocale_DA,
			"DE": N8NDefaultLocale_DE,
			"EL": N8NDefaultLocale_EL,
			"EN": N8NDefaultLocale_EN,
			"ES": N8NDefaultLocale_ES,
			"ET": N8NDefaultLocale_ET,
			"EU": N8NDefaultLocale_EU,
			"FA": N8NDefaultLocale_FA,
			"FI": N8NDefaultLocale_FI,
			"FR": N8NDefaultLocale_FR,
			"GA": N8NDefaultLocale_GA,
			"GL": N8NDefaultLocale_GL,
			"GU": N8NDefaultLocale_GU,
			"HE": N8NDefaultLocale_HE,
			"HI": N8NDefaultLocale_HI,
			"HR": N8NDefaultLocale_HR,
			"HU": N8NDefaultLocale_HU,
			"HY": N8NDefaultLocale_HY,
			"ID": N8NDefaultLocale_ID,
			"IS": N8NDefaultLocale_IS,
			"IT": N8NDefaultLocale_IT,
			"JA": N8NDefaultLocale_JA,
			"KA": N8NDefaultLocale_KA,
			"KK": N8NDefaultLocale_KK,
			"KM": N8NDefaultLocale_KM,
			"KN": N8NDefaultLocale_KN,
			"KO": N8NDefaultLocale_KO,
			"LB": N8NDefaultLocale_LB,
			"LT": N8NDefaultLocale_LT,
			"LV": N8NDefaultLocale_LV,
			"MK": N8NDefaultLocale_MK,
			"ML": N8NDefaultLocale_ML,
			"MR": N8NDefaultLocale_MR,
			"MS": N8NDefaultLocale_MS,
			"MT": N8NDefaultLocale_MT,
			"NB": N8NDefaultLocale_NB,
			"NE": N8NDefaultLocale_NE,
			"NL": N8NDefaultLocale_NL,
			"NN": N8NDefaultLocale_NN,
			"OR": N8NDefaultLocale_OR,
			"PA": N8NDefaultLocale_PA,
			"PL": N8NDefaultLocale_PL,
			"RO": N8NDefaultLocale_RO,
			"RU": N8NDefaultLocale_RU,
			"RW": N8NDefaultLocale_RW,
			"SI": N8NDefaultLocale_SI,
			"SK": N8NDefaultLocale_SK,
			"SL": N8NDefaultLocale_SL,
			"SQ": N8NDefaultLocale_SQ,
			"SV": N8NDefaultLocale_SV,
			"SW": N8NDefaultLocale_SW,
			"TA": N8NDefaultLocale_TA,
			"TE": N8NDefaultLocale_TE,
			"TH": N8NDefaultLocale_TH,
			"TI": N8NDefaultLocale_TI,
			"TN": N8NDefaultLocale_TN,
			"TR": N8NDefaultLocale_TR,
			"UK": N8NDefaultLocale_UK,
			"UR": N8NDefaultLocale_UR,
			"VI": N8NDefaultLocale_VI,
			"WO": N8NDefaultLocale_WO,
			"XH": N8NDefaultLocale_XH,
			"ZU": N8NDefaultLocale_ZU,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NDiagnostics",
		reflect.TypeOf((*N8NDiagnostics)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NDiagnosticsPostHog",
		reflect.TypeOf((*N8NDiagnosticsPostHog)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NDnsConfig",
		reflect.TypeOf((*N8NDnsConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NDnsConfigOptions",
		reflect.TypeOf((*N8NDnsConfigOptions)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NDnsPolicy",
		reflect.TypeOf((*N8NDnsPolicy)(nil)).Elem(),
		map[string]interface{}{
			"CLUSTER_FIRST": N8NDnsPolicy_CLUSTER_FIRST,
			"CLUSTER_FIRST_WITH_HOST_NET": N8NDnsPolicy_CLUSTER_FIRST_WITH_HOST_NET,
			"DEFAULT": N8NDnsPolicy_DEFAULT,
			"NONE": N8NDnsPolicy_NONE,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NExternalPostgresql",
		reflect.TypeOf((*N8NExternalPostgresql)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NExternalRedis",
		reflect.TypeOf((*N8NExternalRedis)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NExternalRedisClusterNodes",
		reflect.TypeOf((*N8NExternalRedisClusterNodes)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NExternalRedisTls",
		reflect.TypeOf((*N8NExternalRedisTls)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NImage",
		reflect.TypeOf((*N8NImage)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NIngress",
		reflect.TypeOf((*N8NIngress)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NLicense",
		reflect.TypeOf((*N8NLicense)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NLicenseAutoNenew",
		reflect.TypeOf((*N8NLicenseAutoNenew)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NLivenessProbe",
		reflect.TypeOf((*N8NLivenessProbe)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NLivenessProbeHttpGet",
		reflect.TypeOf((*N8NLivenessProbeHttpGet)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NLog",
		reflect.TypeOf((*N8NLog)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NLogFile",
		reflect.TypeOf((*N8NLogFile)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NLogLevel",
		reflect.TypeOf((*N8NLogLevel)(nil)).Elem(),
		map[string]interface{}{
			"ERROR": N8NLogLevel_ERROR,
			"WARN": N8NLogLevel_WARN,
			"INFO": N8NLogLevel_INFO,
			"DEBUG": N8NLogLevel_DEBUG,
		},
	)
	_jsii_.RegisterEnum(
		"n8n.N8NLogOutput",
		reflect.TypeOf((*N8NLogOutput)(nil)).Elem(),
		map[string]interface{}{
			"CONSOLE": N8NLogOutput_CONSOLE,
			"FILE": N8NLogOutput_FILE,
		},
	)
	_jsii_.RegisterEnum(
		"n8n.N8NLogScopes",
		reflect.TypeOf((*N8NLogScopes)(nil)).Elem(),
		map[string]interface{}{
			"CONCURRENCY": N8NLogScopes_CONCURRENCY,
			"EXTERNAL_HYPHEN_SECRETS": N8NLogScopes_EXTERNAL_HYPHEN_SECRETS,
			"LICENSE": N8NLogScopes_LICENSE,
			"MULTI_HYPHEN_MAIN_HYPHEN_SETUP": N8NLogScopes_MULTI_HYPHEN_MAIN_HYPHEN_SETUP,
			"PUBSUB": N8NLogScopes_PUBSUB,
			"REDIS": N8NLogScopes_REDIS,
			"SCALING": N8NLogScopes_SCALING,
			"WAITING_HYPHEN_EXECUTIONS": N8NLogScopes_WAITING_HYPHEN_EXECUTIONS,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NMain",
		reflect.TypeOf((*N8NMain)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NMainHostAliases",
		reflect.TypeOf((*N8NMainHostAliases)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NMainLivenessProbe",
		reflect.TypeOf((*N8NMainLivenessProbe)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NMainLivenessProbeHttpGet",
		reflect.TypeOf((*N8NMainLivenessProbeHttpGet)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NMainPdb",
		reflect.TypeOf((*N8NMainPdb)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NMainPersistence",
		reflect.TypeOf((*N8NMainPersistence)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NMainPersistenceAccessMode",
		reflect.TypeOf((*N8NMainPersistenceAccessMode)(nil)).Elem(),
		map[string]interface{}{
			"READ_WRITE_ONCE": N8NMainPersistenceAccessMode_READ_WRITE_ONCE,
			"READ_WRITE_MANY": N8NMainPersistenceAccessMode_READ_WRITE_MANY,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NMainReadinessProbe",
		reflect.TypeOf((*N8NMainReadinessProbe)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NMainReadinessProbeHttpGet",
		reflect.TypeOf((*N8NMainReadinessProbeHttpGet)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NMainResources",
		reflect.TypeOf((*N8NMainResources)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NMainResourcesLimits",
		reflect.TypeOf((*N8NMainResourcesLimits)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NMainResourcesRequests",
		reflect.TypeOf((*N8NMainResourcesRequests)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NNodes",
		reflect.TypeOf((*N8NNodes)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NNodesBuiltin",
		reflect.TypeOf((*N8NNodesBuiltin)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NNodesExternal",
		reflect.TypeOf((*N8NNodesExternal)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NNodesInitContainer",
		reflect.TypeOf((*N8NNodesInitContainer)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NNodesInitContainerImage",
		reflect.TypeOf((*N8NNodesInitContainerImage)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NNodesInitContainerResources",
		reflect.TypeOf((*N8NNodesInitContainerResources)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NNodesInitContainerResourcesLimits",
		reflect.TypeOf((*N8NNodesInitContainerResourcesLimits)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NNodesInitContainerResourcesRequests",
		reflect.TypeOf((*N8NNodesInitContainerResourcesRequests)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NNpmRegistry",
		reflect.TypeOf((*N8NNpmRegistry)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NPodSecurityContext",
		reflect.TypeOf((*N8NPodSecurityContext)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NPodSecurityContextFsGroupChangePolicy",
		reflect.TypeOf((*N8NPodSecurityContextFsGroupChangePolicy)(nil)).Elem(),
		map[string]interface{}{
			"ON_ROOT_MISMATCH": N8NPodSecurityContextFsGroupChangePolicy_ON_ROOT_MISMATCH,
			"ALWAYS": N8NPodSecurityContextFsGroupChangePolicy_ALWAYS,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NReadinessProbe",
		reflect.TypeOf((*N8NReadinessProbe)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NReadinessProbeHttpGet",
		reflect.TypeOf((*N8NReadinessProbeHttpGet)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NResources",
		reflect.TypeOf((*N8NResources)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NResourcesLimits",
		reflect.TypeOf((*N8NResourcesLimits)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NResourcesRequests",
		reflect.TypeOf((*N8NResourcesRequests)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NSecurityContext",
		reflect.TypeOf((*N8NSecurityContext)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NSecurityContextCapabilities",
		reflect.TypeOf((*N8NSecurityContextCapabilities)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NSentry",
		reflect.TypeOf((*N8NSentry)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NService",
		reflect.TypeOf((*N8NService)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NServiceAccount",
		reflect.TypeOf((*N8NServiceAccount)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NServiceMonitor",
		reflect.TypeOf((*N8NServiceMonitor)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NServiceMonitorInclude",
		reflect.TypeOf((*N8NServiceMonitorInclude)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NServiceMonitorMetricRelabelings",
		reflect.TypeOf((*N8NServiceMonitorMetricRelabelings)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NServiceMonitorMetricRelabelingsAction",
		reflect.TypeOf((*N8NServiceMonitorMetricRelabelingsAction)(nil)).Elem(),
		map[string]interface{}{
			"REPLACE": N8NServiceMonitorMetricRelabelingsAction_REPLACE,
			"KEEP": N8NServiceMonitorMetricRelabelingsAction_KEEP,
			"DROP": N8NServiceMonitorMetricRelabelingsAction_DROP,
			"LABELDROP": N8NServiceMonitorMetricRelabelingsAction_LABELDROP,
			"LABELKEEP": N8NServiceMonitorMetricRelabelingsAction_LABELKEEP,
			"HASHMOD": N8NServiceMonitorMetricRelabelingsAction_HASHMOD,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NStrategy",
		reflect.TypeOf((*N8NStrategy)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NStrategyType",
		reflect.TypeOf((*N8NStrategyType)(nil)).Elem(),
		map[string]interface{}{
			"ROLLING_UPDATE": N8NStrategyType_ROLLING_UPDATE,
			"RECREATE": N8NStrategyType_RECREATE,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NTaskRunners",
		reflect.TypeOf((*N8NTaskRunners)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NTaskRunnersBroker",
		reflect.TypeOf((*N8NTaskRunnersBroker)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NTaskRunnersExternal",
		reflect.TypeOf((*N8NTaskRunnersExternal)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NTaskRunnersExternalResources",
		reflect.TypeOf((*N8NTaskRunnersExternalResources)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NTaskRunnersExternalResourcesLimits",
		reflect.TypeOf((*N8NTaskRunnersExternalResourcesLimits)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NTaskRunnersExternalResourcesRequests",
		reflect.TypeOf((*N8NTaskRunnersExternalResourcesRequests)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NTaskRunnersMode",
		reflect.TypeOf((*N8NTaskRunnersMode)(nil)).Elem(),
		map[string]interface{}{
			"INTERNAL": N8NTaskRunnersMode_INTERNAL,
			"EXTERNAL": N8NTaskRunnersMode_EXTERNAL,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NVersionNotifications",
		reflect.TypeOf((*N8NVersionNotifications)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhook",
		reflect.TypeOf((*N8NWebhook)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookAutoscaling",
		reflect.TypeOf((*N8NWebhookAutoscaling)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookAutoscalingBehavior",
		reflect.TypeOf((*N8NWebhookAutoscalingBehavior)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookAutoscalingBehaviorScaleDown",
		reflect.TypeOf((*N8NWebhookAutoscalingBehaviorScaleDown)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookAutoscalingBehaviorScaleDownPolicies",
		reflect.TypeOf((*N8NWebhookAutoscalingBehaviorScaleDownPolicies)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NWebhookAutoscalingBehaviorScaleDownPoliciesType",
		reflect.TypeOf((*N8NWebhookAutoscalingBehaviorScaleDownPoliciesType)(nil)).Elem(),
		map[string]interface{}{
			"PODS": N8NWebhookAutoscalingBehaviorScaleDownPoliciesType_PODS,
			"PERCENT": N8NWebhookAutoscalingBehaviorScaleDownPoliciesType_PERCENT,
		},
	)
	_jsii_.RegisterEnum(
		"n8n.N8NWebhookAutoscalingBehaviorScaleDownSelectPolicy",
		reflect.TypeOf((*N8NWebhookAutoscalingBehaviorScaleDownSelectPolicy)(nil)).Elem(),
		map[string]interface{}{
			"MAX": N8NWebhookAutoscalingBehaviorScaleDownSelectPolicy_MAX,
			"MIN": N8NWebhookAutoscalingBehaviorScaleDownSelectPolicy_MIN,
			"DISABLED": N8NWebhookAutoscalingBehaviorScaleDownSelectPolicy_DISABLED,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookAutoscalingBehaviorScaleUp",
		reflect.TypeOf((*N8NWebhookAutoscalingBehaviorScaleUp)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookAutoscalingBehaviorScaleUpPolicies",
		reflect.TypeOf((*N8NWebhookAutoscalingBehaviorScaleUpPolicies)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NWebhookAutoscalingBehaviorScaleUpPoliciesType",
		reflect.TypeOf((*N8NWebhookAutoscalingBehaviorScaleUpPoliciesType)(nil)).Elem(),
		map[string]interface{}{
			"PODS": N8NWebhookAutoscalingBehaviorScaleUpPoliciesType_PODS,
			"PERCENT": N8NWebhookAutoscalingBehaviorScaleUpPoliciesType_PERCENT,
		},
	)
	_jsii_.RegisterEnum(
		"n8n.N8NWebhookAutoscalingBehaviorScaleUpSelectPolicy",
		reflect.TypeOf((*N8NWebhookAutoscalingBehaviorScaleUpSelectPolicy)(nil)).Elem(),
		map[string]interface{}{
			"MAX": N8NWebhookAutoscalingBehaviorScaleUpSelectPolicy_MAX,
			"MIN": N8NWebhookAutoscalingBehaviorScaleUpSelectPolicy_MIN,
			"DISABLED": N8NWebhookAutoscalingBehaviorScaleUpSelectPolicy_DISABLED,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookAutoscalingMetrics",
		reflect.TypeOf((*N8NWebhookAutoscalingMetrics)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NWebhookAutoscalingMetricsType",
		reflect.TypeOf((*N8NWebhookAutoscalingMetricsType)(nil)).Elem(),
		map[string]interface{}{
			"RESOURCE": N8NWebhookAutoscalingMetricsType_RESOURCE,
			"PODS": N8NWebhookAutoscalingMetricsType_PODS,
			"OBJECT": N8NWebhookAutoscalingMetricsType_OBJECT,
			"EXTERNAL": N8NWebhookAutoscalingMetricsType_EXTERNAL,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookHostAliases",
		reflect.TypeOf((*N8NWebhookHostAliases)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookLivenessProbe",
		reflect.TypeOf((*N8NWebhookLivenessProbe)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookLivenessProbeHttpGet",
		reflect.TypeOf((*N8NWebhookLivenessProbeHttpGet)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookMcp",
		reflect.TypeOf((*N8NWebhookMcp)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookMcpHostAliases",
		reflect.TypeOf((*N8NWebhookMcpHostAliases)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookMcpLivenessProbe",
		reflect.TypeOf((*N8NWebhookMcpLivenessProbe)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookMcpLivenessProbeHttpGet",
		reflect.TypeOf((*N8NWebhookMcpLivenessProbeHttpGet)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookMcpReadinessProbe",
		reflect.TypeOf((*N8NWebhookMcpReadinessProbe)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookMcpReadinessProbeHttpGet",
		reflect.TypeOf((*N8NWebhookMcpReadinessProbeHttpGet)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookMcpResources",
		reflect.TypeOf((*N8NWebhookMcpResources)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookMcpResourcesLimits",
		reflect.TypeOf((*N8NWebhookMcpResourcesLimits)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookMcpResourcesRequests",
		reflect.TypeOf((*N8NWebhookMcpResourcesRequests)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookMcpStartupProbe",
		reflect.TypeOf((*N8NWebhookMcpStartupProbe)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookMcpStartupProbeExec",
		reflect.TypeOf((*N8NWebhookMcpStartupProbeExec)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NWebhookMode",
		reflect.TypeOf((*N8NWebhookMode)(nil)).Elem(),
		map[string]interface{}{
			"REGULAR": N8NWebhookMode_REGULAR,
			"QUEUE": N8NWebhookMode_QUEUE,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookPdb",
		reflect.TypeOf((*N8NWebhookPdb)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookReadinessProbe",
		reflect.TypeOf((*N8NWebhookReadinessProbe)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookReadinessProbeHttpGet",
		reflect.TypeOf((*N8NWebhookReadinessProbeHttpGet)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookResources",
		reflect.TypeOf((*N8NWebhookResources)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookResourcesLimits",
		reflect.TypeOf((*N8NWebhookResourcesLimits)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookResourcesRequests",
		reflect.TypeOf((*N8NWebhookResourcesRequests)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookStartupProbe",
		reflect.TypeOf((*N8NWebhookStartupProbe)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookStartupProbeExec",
		reflect.TypeOf((*N8NWebhookStartupProbeExec)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWebhookWaitMainNodeReady",
		reflect.TypeOf((*N8NWebhookWaitMainNodeReady)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorker",
		reflect.TypeOf((*N8NWorker)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerAutoscaling",
		reflect.TypeOf((*N8NWorkerAutoscaling)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerAutoscalingBehavior",
		reflect.TypeOf((*N8NWorkerAutoscalingBehavior)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerAutoscalingBehaviorScaleDown",
		reflect.TypeOf((*N8NWorkerAutoscalingBehaviorScaleDown)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerAutoscalingBehaviorScaleDownPolicies",
		reflect.TypeOf((*N8NWorkerAutoscalingBehaviorScaleDownPolicies)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NWorkerAutoscalingBehaviorScaleDownPoliciesType",
		reflect.TypeOf((*N8NWorkerAutoscalingBehaviorScaleDownPoliciesType)(nil)).Elem(),
		map[string]interface{}{
			"PODS": N8NWorkerAutoscalingBehaviorScaleDownPoliciesType_PODS,
			"PERCENT": N8NWorkerAutoscalingBehaviorScaleDownPoliciesType_PERCENT,
		},
	)
	_jsii_.RegisterEnum(
		"n8n.N8NWorkerAutoscalingBehaviorScaleDownSelectPolicy",
		reflect.TypeOf((*N8NWorkerAutoscalingBehaviorScaleDownSelectPolicy)(nil)).Elem(),
		map[string]interface{}{
			"MAX": N8NWorkerAutoscalingBehaviorScaleDownSelectPolicy_MAX,
			"MIN": N8NWorkerAutoscalingBehaviorScaleDownSelectPolicy_MIN,
			"DISABLED": N8NWorkerAutoscalingBehaviorScaleDownSelectPolicy_DISABLED,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerAutoscalingBehaviorScaleUp",
		reflect.TypeOf((*N8NWorkerAutoscalingBehaviorScaleUp)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerAutoscalingBehaviorScaleUpPolicies",
		reflect.TypeOf((*N8NWorkerAutoscalingBehaviorScaleUpPolicies)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NWorkerAutoscalingBehaviorScaleUpPoliciesType",
		reflect.TypeOf((*N8NWorkerAutoscalingBehaviorScaleUpPoliciesType)(nil)).Elem(),
		map[string]interface{}{
			"PODS": N8NWorkerAutoscalingBehaviorScaleUpPoliciesType_PODS,
			"PERCENT": N8NWorkerAutoscalingBehaviorScaleUpPoliciesType_PERCENT,
		},
	)
	_jsii_.RegisterEnum(
		"n8n.N8NWorkerAutoscalingBehaviorScaleUpSelectPolicy",
		reflect.TypeOf((*N8NWorkerAutoscalingBehaviorScaleUpSelectPolicy)(nil)).Elem(),
		map[string]interface{}{
			"MAX": N8NWorkerAutoscalingBehaviorScaleUpSelectPolicy_MAX,
			"MIN": N8NWorkerAutoscalingBehaviorScaleUpSelectPolicy_MIN,
			"DISABLED": N8NWorkerAutoscalingBehaviorScaleUpSelectPolicy_DISABLED,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerAutoscalingMetrics",
		reflect.TypeOf((*N8NWorkerAutoscalingMetrics)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NWorkerAutoscalingMetricsType",
		reflect.TypeOf((*N8NWorkerAutoscalingMetricsType)(nil)).Elem(),
		map[string]interface{}{
			"RESOURCE": N8NWorkerAutoscalingMetricsType_RESOURCE,
			"PODS": N8NWorkerAutoscalingMetricsType_PODS,
			"OBJECT": N8NWorkerAutoscalingMetricsType_OBJECT,
			"EXTERNAL": N8NWorkerAutoscalingMetricsType_EXTERNAL,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerHostAliases",
		reflect.TypeOf((*N8NWorkerHostAliases)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerLivenessProbe",
		reflect.TypeOf((*N8NWorkerLivenessProbe)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerLivenessProbeHttpGet",
		reflect.TypeOf((*N8NWorkerLivenessProbeHttpGet)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NWorkerMode",
		reflect.TypeOf((*N8NWorkerMode)(nil)).Elem(),
		map[string]interface{}{
			"REGULAR": N8NWorkerMode_REGULAR,
			"QUEUE": N8NWorkerMode_QUEUE,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerPdb",
		reflect.TypeOf((*N8NWorkerPdb)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerPersistence",
		reflect.TypeOf((*N8NWorkerPersistence)(nil)).Elem(),
	)
	_jsii_.RegisterEnum(
		"n8n.N8NWorkerPersistenceAccessMode",
		reflect.TypeOf((*N8NWorkerPersistenceAccessMode)(nil)).Elem(),
		map[string]interface{}{
			"READ_WRITE_ONCE": N8NWorkerPersistenceAccessMode_READ_WRITE_ONCE,
			"READ_WRITE_MANY": N8NWorkerPersistenceAccessMode_READ_WRITE_MANY,
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerReadinessProbe",
		reflect.TypeOf((*N8NWorkerReadinessProbe)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerReadinessProbeHttpGet",
		reflect.TypeOf((*N8NWorkerReadinessProbeHttpGet)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerResources",
		reflect.TypeOf((*N8NWorkerResources)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerResourcesLimits",
		reflect.TypeOf((*N8NWorkerResourcesLimits)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerResourcesRequests",
		reflect.TypeOf((*N8NWorkerResourcesRequests)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerStartupProbe",
		reflect.TypeOf((*N8NWorkerStartupProbe)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerStartupProbeExec",
		reflect.TypeOf((*N8NWorkerStartupProbeExec)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkerWaitMainNodeReady",
		reflect.TypeOf((*N8NWorkerWaitMainNodeReady)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8NWorkflowHistory",
		reflect.TypeOf((*N8NWorkflowHistory)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"n8n.N8n",
		reflect.TypeOf((*N8n)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "helm", GoGetter: "Helm"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "with", GoMethod: "With"},
		},
		func() interface{} {
			j := jsiiProxy_N8n{}
			_jsii_.InitJsiiProxy(&j.Type__constructsConstruct)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"n8n.N8nProps",
		reflect.TypeOf((*N8nProps)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"n8n.N8nValues",
		reflect.TypeOf((*N8nValues)(nil)).Elem(),
	)
}

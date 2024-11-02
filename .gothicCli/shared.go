package gothicCliShared

type Config struct {
	ProjectName    string `json:"projectName"`
	OptimizeImages struct {
		LowResolutionRate int `json:"lowResolutionRate"`
	} `json:"optimizeImages"`
	Deploy *DeployConfig `json:"deploy"`
}

type DeployConfig struct {
	ServerMemory  int               `json:"serverMemory"`
	ServerTimeout int               `json:"serverTimeout"`
	Region        string            `json:"region"`
	Stages        EnvironmentConfig `json:"stages"`
	CustomDomain  bool              `json:"customDomain"`
}
type EnvVariables struct {
	BucketName     string                 `json:"BucketName"`
	LambdaName     string                 `json:"LambdaName"`
	HostedZoneId   *string                `json:"hostedZoneId"`
	CustomDomain   *string                `json:"customDomain"`
	CertificateArn *string                `json:"certificateArn"`
	ENV            map[string]interface{} `json:"env,omitempty"`
}

type EnvironmentConfig struct {
	Default EnvVariables `json:"default"`
	Dev     EnvVariables `json:"dev"`
	Staging EnvVariables `json:"staging"`
	Prod    EnvVariables `json:"prod"`
}

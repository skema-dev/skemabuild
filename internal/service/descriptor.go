package service

type ServiceMethodDescriptor struct {
	Name          string
	NameCamelCase string
	RequestType   string
	ResponseType  string
}

type ServiceDescriptor struct {
	Name    string
	Methods []ServiceMethodDescriptor
}

type RpcParameters struct {
	GoModule             string
	GoVersion            string
	GoPackageAddress     string
	HttpEnabled          bool
	ServiceName          string
	ServiceNameCamelCase string
	ServiceNameLower     string

	RpcServices []ServiceDescriptor
}

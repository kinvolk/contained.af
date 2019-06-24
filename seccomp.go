package main

var seccompConfigs = map[dockerProfile]string{
	defaultDockerProfile: defaultSeccompConfig,
	weakDockerProfile:    weakSeccompConfig,
}

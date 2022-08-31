package main

import (
	"errors"
	"fmt"
	"github.com/nabsul/k8s-ecr-login-renew/src/aws"
	"github.com/nabsul/k8s-ecr-login-renew/src/k8s"
	"os"
	"strings"
	"time"
)

const (
	envVarAwsSecret       = "DOCKER_SECRET_NAME"
	envVarTargetNamespace = "TARGET_NAMESPACE"
	envVarRegistries      = "DOCKER_REGISTRIES"
	envVarAnnotations     = "ANNOTATIONS"
	envVarLabels          = "LABELS"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println("Running at " + time.Now().UTC().String())

	name := os.Getenv(envVarAwsSecret)
	if name == "" {
		panic(fmt.Sprintf("Environment variable %s is required", envVarAwsSecret))
	}

	annotationsString, ok := os.LookupEnv(envVarAnnotations)
	if !ok {
		annotationsString = ""
	}
	annotations := stringToMap(annotationsString)
	fmt.Printf("Annotations to appy: %+v\n", annotations)

	labelsString, ok := os.LookupEnv(envVarLabels)
	if !ok {
		labelsString = ""
	}
	labels := stringToMap(labelsString)
	fmt.Printf("Labels to appy: %+v\n", labels)

	fmt.Print("Fetching auth data from AWS... ")
	credentials, err := aws.GetDockerCredentials()
	checkErr(err)
	fmt.Println("Success.")

	servers := getServerList(credentials.Server)
	fmt.Printf("Docker Registries: %s\n", strings.Join(servers, ","))

	namespaces, err := k8s.GetNamespaces(os.Getenv(envVarTargetNamespace))
	checkErr(err)
	fmt.Printf("Updating kubernetes secret [%s] in %d namespaces\n", name, len(namespaces))

	failed := false
	for _, ns := range namespaces {
		fmt.Printf("Updating secret in namespace [%s]... ", ns)
		err = k8s.UpdatePassword(ns, name, credentials.Username, credentials.Password, servers, annotations, labels)
		if nil != err {
			fmt.Printf("failed: %s\n", err)
			failed = true
		} else {
			fmt.Println("success")
		}
	}

	if failed {
		panic(errors.New("failed to create one of more Docker login secrets"))
	}

	fmt.Println("Job complete.")
}

func getServerList(defaultServer string) []string {
	addedServersSetting := os.Getenv(envVarRegistries)

	if addedServersSetting == "" {
		return []string{defaultServer}
	}

	return strings.Split(addedServersSetting, ",")
}

func stringToMap(str string) map[string]string {
	m := map[string]string{}
	if len(str) == 0 {
		return m
	}

	for _, item := range strings.Split(str, ",") {
		itemSlice := strings.Split(item, "=")
		m[itemSlice[0]] = itemSlice[1]
	}

	return m
}

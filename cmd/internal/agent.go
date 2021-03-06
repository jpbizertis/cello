/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package internal

import (
	"errors"
	"fmt"
	ghYaml "github.com/ghodss/yaml"
	"github.com/hyperledger/cello/cmd/model"
	"github.com/levigross/grequests"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"log"
)

var agentBaseUrl = "agents"

func ListAgents(baseUrl string, token string) (*model.AgentListResponse, error) {
	url := fmt.Sprintf("%s/%s", baseUrl, agentBaseUrl)
	resp, err := grequests.Get(url, &grequests.RequestOptions{Headers: map[string]string{
		"Authorization": token,
	}})

	if err != nil {
		log.Fatalln("Unable to make request: ", err)
		return nil, errors.New("list agents failed")
	}
	if resp.Ok != true {
		log.Printf("Get agents failed %v %v", resp.StatusCode, resp.String())
		return nil, errors.New("list agents failed")
	} else {
		var agentListResponse model.AgentListResponse
		err := resp.JSON(&agentListResponse)
		if err != nil {
			panic(err)
		}

		return &agentListResponse, nil
	}
}

func parseAgentSpec(source [] byte) (*model.AgentSpec, error) {
	var agentSpec model.AgentSpec
	err := yaml.Unmarshal(source, &agentSpec)
	if err != nil {
		return nil, err
	} else {
		return &agentSpec, nil
	}
}

func CreateAgent(source []byte) error {
	agentSpec, err := parseAgentSpec(source)
	if err != nil {
		panic(err)
	}
	yamlOuts, err := yaml.Marshal(agentSpec.Spec)
	if err != nil {
		panic(err)
	}
	jsonOuts, err := ghYaml.YAMLToJSON(yamlOuts)
	if err != nil {
		panic(err)
	}

	url := fmt.Sprintf("%s/%s", viper.GetString("server.url"), agentBaseUrl)
	token := fmt.Sprintf("JWT %s", viper.GetString("auth.token"))

	switch agentSpec.Spec.Type {
	case "docker":
		resp, err := grequests.Post(url, &grequests.RequestOptions{JSON: jsonOuts, Headers: map[string] string {
			"Authorization": token,
		}})
		if err != nil {
			fmt.Printf("Create agent failed")
			return err
		} else {
			fmt.Printf(resp.String())
		}
		break
	case "kubernetes":
		fd, err := grequests.FileUploadFromDisk(agentSpec.Spec.K8sConfigFile)
		if err != nil {
			panic(err)
		}
		fd[0].FieldName = "k8s_config_file"
		x := map[string] string {
			"name": agentSpec.Spec.Name,
			"type": agentSpec.Spec.Type,
			"capacity": fmt.Sprint(agentSpec.Spec.Capacity),
			"log_level": agentSpec.Spec.LogLevel,
			"node_capacity": fmt.Sprint(agentSpec.Spec.NodeCapacity),
			"schedulable": fmt.Sprint(agentSpec.Spec.Schedulable),
		}
		resp, err := grequests.Post(url,
			&grequests.RequestOptions{
				Files: fd,
				Data: x,
				Headers: map[string] string {
					"Authorization": token,
				},
			})
		if err != nil {
			fmt.Printf("Create agent failed")
			return err
		} else {
			fmt.Printf(resp.String())
		}
		break
	default:
		fmt.Printf("Unknow type for agent")
		return nil
	}
	return nil
}

func DeleteAgent(id string) error {
	url := fmt.Sprintf("%s/%s/%s", viper.GetString("server.url"), agentBaseUrl, id)
	token := fmt.Sprintf("JWT %s", viper.GetString("auth.token"))
	resp, err := grequests.Delete(url, &grequests.RequestOptions{Headers: map[string] string {
		"Authorization": token,
	}})
	if err != nil {
		return err
	} else {
		if resp.StatusCode != 204 {
			errMsg := fmt.Sprintf("delete agent %s failed, %s", id, resp.String())
			return errors.New(errMsg)
		} else {
			fmt.Printf("Delete agent %s success\n", id)
		}
	}
	return nil
}

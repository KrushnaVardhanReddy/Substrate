package gateway

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"sigs.k8s.io/yaml"
)

type KongGenerator struct{}

type KongIngress struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name        string            `json:"name"`
		Annotations map[string]string `json:"annotations,omitempty"`
	} `json:"metadata"`
	Proxy struct {
		Paths []string `json:"paths"`
	} `json:"proxy"`
}

type Ingress struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name        string            `json:"name"`
		Annotations map[string]string `json:"annotations,omitempty"`
	} `json:"metadata"`
	Spec struct {
		Rules []IngressRule `json:"rules"`
	} `json:"spec"`
}

type IngressRule struct {
	Http struct {
		Paths []IngressPath `json:"paths"`
	} `json:"http"`
}

type IngressPath struct {
	Path     string `json:"path"`
	PathType string `json:"pathType"`
	Backend  struct {
		Service struct {
			Name string `json:"name"`
			Port struct {
				Number int `json:"number"`
			} `json:"port"`
		} `json:"service"`
	} `json:"backend"`
}

func (g *KongGenerator) GenerateCRD(spec *openapi3.T, repo string) (string, error) {
	if spec == nil {
		return "", fmt.Errorf("openapi spec is nil")
	}

	kongIngress := KongIngress{
		ApiVersion: "configuration.konghq.com/v1",
		Kind:       "KongIngress",
	}
	kongIngress.Metadata.Name = repo
	kongIngress.Proxy.Paths = make([]string, 0)

	ingress := Ingress{
		ApiVersion: "networking.k8s.io/v1",
		Kind:       "Ingress",
	}
	ingress.Metadata.Name = repo
	ingress.Metadata.Annotations = map[string]string{
		"konghq.com/override": repo,
	}

	paths := make([]IngressPath, 0)

	if spec.Paths != nil {
		for path := range spec.Paths.Map() {
			kongIngress.Proxy.Paths = append(kongIngress.Proxy.Paths, path)
			paths = append(paths, IngressPath{
				Path:     path,
				PathType: "Prefix",
				Backend: struct {
					Service struct {
						Name string `json:"name"`
						Port struct {
							Number int `json:"number"`
						} `json:"port"`
					} `json:"service"`
				}{
					Service: struct {
						Name string `json:"name"`
						Port struct {
							Number int `json:"number"`
						} `json:"port"`
					}{
						Name: repo,
						Port: struct {
							Number int `json:"number"`
						}{Number: 80},
					},
				},
			})
		}
	}

	ingress.Spec.Rules = []IngressRule{
		{
			Http: struct {
				Paths []IngressPath `json:"paths"`
			}{
				Paths: paths,
			},
		},
	}

	kongYaml, err := yaml.Marshal(kongIngress)
	if err != nil {
		return "", err
	}

	ingressYaml, err := yaml.Marshal(ingress)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("---\n%s---\n%s", string(kongYaml), string(ingressYaml)), nil
}

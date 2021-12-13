package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/EdgeNet-project/fed4fire/pkg/identifiers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type PerformOperationalActionArgs struct {
	URNs        []string
	Credentials []Credential
	Action      string
	// The only supported action is `geni_update_users`.
	Options UpdateUsersOptions
}

type PerformOperationalActionReply struct {
	Data struct {
		Code  Code     `xml:"code"`
		Value []Sliver `xml:"value"`
	}
}

type UpdateUsersOptions struct {
	Users []struct {
		URN  string   `xml:"urn"`
		Keys []string `xml:"keys"`
	} `xml:"geni_users"`
}

// PerformOperationalAction performs the named operational action on the named slivers,
// possibly changing the geni_operational_status of the named slivers, e.g. 'start' a VM.
// For valid operations and expected states, consult the state diagram advertised in the aggregate's advertisement RSpec.
func (s *Service) PerformOperationalAction(
	r *http.Request,
	args *PerformOperationalActionArgs,
	reply *PerformOperationalActionReply,
) error {
	// TODO: Check credentials
	if args.Action != "geni_update_users" {
		// TODO: Handle error
		return nil
	}
	keys := make([]string, 0)
	for _, user := range args.Options.Users {
		for _, key := range user.Keys {
			keys = append(keys, key)
		}
	}
	configMaps := make([]corev1.ConfigMap, 0)
	pods := make([]corev1.Pod, 0)
	for _, urn := range args.URNs {
		identifier, err := identifiers.Parse(urn)
		if err != nil {
			// TODO: Handle error
			fmt.Println(err)
		}
		configMaps_, err := s.GetConfigMaps(r.Context(), *identifier)
		if err != nil {
			// TODO: Handle error
			fmt.Println(err)
		}
		configMaps = append(configMaps, configMaps_...)
		pods_, err := s.GetPods(r.Context(), *identifier)
		if err != nil {
			// TODO: Handle error
			fmt.Println(err)
		}
		pods = append(pods, pods_...)
	}
	for _, configMap := range configMaps {
		configMap.Data = map[string]string{
			"authorized_keys": strings.Join(keys, "\n") + "\n",
		}
		_, err := s.ConfigMaps().Update(r.Context(), &configMap, metav1.UpdateOptions{})
		if err == nil {
			klog.InfoS("Updated configmap", "name", configMap.Name)
		} else {
			// TODO: Handle error
			fmt.Println(err)
		}
	}
	// TODO: Explain why we do this
	// TODO: Delete pod only if keys have changed?
	for _, pod := range pods {
		err := s.Pods().Delete(r.Context(), pod.Name, metav1.DeleteOptions{})
		if err == nil {
			klog.InfoS("Deleted pod", "name", pod.Name)
		} else {
			// TODO: Handle error
			fmt.Println(err)
		}
	}
	return nil
}

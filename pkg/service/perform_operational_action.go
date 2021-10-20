package service

import "net/http"

type PerformOperationalActionArgs struct {
	URNs        []string
	Credentials []Credential
	Action      string
	Options     string // Options
}

type PerformOperationalActionReply struct {
	Data struct {
		Code  Code     `xml:"code"`
		Value []Sliver `xml:"value"`
	}
}

//geni_update_users
//{
//   "geni_best_effort": false,
//   "geni_users": [
//          {
//             "urn": "urn:publicid:IDN+ilabt.imec.be+user+mouchetm",
//             "keys": [
//                          "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC0F+kIOsaIfSi3fX1xEBE/sF2WutXKqcUkquP+IbY8OC1TbwOzZtrwX6k64gmCxW2mfoLgsxzSVuj9mu5X9xv/j1DYd92TNxCj7EOy/cQa2hCZpXW2feGKbsVY1KyEn91zDeRlIovtrKMvg7F4tXZlEruLA6uOmPLaCOS55UrgVOqQXIZyZYoK1AL3srSdNuIf5/vB61fKHUsfwbNjxcCdl12YL4gxFBlO/QccAjkUVRYUkkIdlNN/BoV41znQ1W0Zy+k/xOgE7LxqP6n8bWSgeyZYOw5jjlVzg1J6074TedStxkJZhoCkbvneRer0UlYFw8ZYtY7LHTDMh4db2YGP"
//                          ]
//             }
//          ]
//   }

// PerformOperationalAction performs the named operational action on the named slivers,
// possibly changing the geni_operational_status of the named slivers, e.g. 'start' a VM.
// For valid operations and expected states, consult the state diagram advertised in the aggregate's advertisement RSpec.
func (s *Service) PerformOperationalAction(
	r *http.Request,
	args *PerformOperationalActionArgs,
	reply *PerformOperationalActionReply,
) error {
	//s.Deployments().Update(r.Context(), )
	return nil
}

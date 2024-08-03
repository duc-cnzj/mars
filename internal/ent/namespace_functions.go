package ent

import (
	"github.com/duc-cnzj/mars/api/v4/types"
)

func (ns *Namespace) GetImagePullSecrets() []*types.ImagePullSecret {
	var secrets = make([]*types.ImagePullSecret, 0)
	for _, s := range ns.ImagePullSecrets {
		secrets = append(secrets, &types.ImagePullSecret{Name: s})
	}
	return secrets
}

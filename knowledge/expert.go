package knowledge

import "github.com/pingcap/ticp/knowledge/models"

var SpecKnowledge *ClusterSpecKnowledge

type ClusterSpecKnowledge struct {
	Specs 		[]*models.ClusterTypeSpec
	Types 		map[string]*models.ClusterType
	Versions 	map[string]*models.ClusterVersion
	Components  map[string]*models.ClusterComponent
}

func ClusterTypeFromCode(code string) *models.ClusterType {
	return SpecKnowledge.Types[code]
}

func ClusterVersionFromCode(code string) *models.ClusterVersion {
	return SpecKnowledge.Versions[code]
}

func ClusterComponentFromCode(componentType string) *models.ClusterComponent {
	return SpecKnowledge.Components[componentType]
}

func SortedTypesKnowledge() []*models.ClusterTypeSpec {
	slice := make([]*models.ClusterTypeSpec, len(SpecKnowledge.Specs), len(SpecKnowledge.Specs))
	for _, v := range SpecKnowledge.Specs  {
		slice = append(slice, v)
	}

	return slice
}

func LoadSpecKnowledge() {

}
package knowledge

var SpecKnowledge *ClusterSpecKnowledge

type ClusterSpecKnowledge struct {
	Specs 		[]*ClusterTypeSpec
	Types 		map[string]*ClusterType
	Versions 	map[string]*ClusterVersion
	Components  map[string]*ClusterComponent
}

func ClusterTypeFromCode(code string) *ClusterType {
	return SpecKnowledge.Types[code]
}

func ClusterVersionFromCode(code string) *ClusterVersion {
	return SpecKnowledge.Versions[code]
}

func ClusterComponentFromCode(componentType string) *ClusterComponent {
	return SpecKnowledge.Components[componentType]
}

func SortedTypesKnowledge() []*ClusterTypeSpec {
	slice := make([]*ClusterTypeSpec, len(SpecKnowledge.Specs), len(SpecKnowledge.Specs))
	for _, v := range SpecKnowledge.Specs  {
		slice = append(slice, v)
	}

	return slice
}

func LoadSpecKnowledge() {

}
package knowledge

var SpecKnowledge *ClusterSpecKnowledge

type ClusterSpecKnowledge struct {
	Specs 		[]*ClusterTypeSpec
	Types 		map[string]*ClusterType
	Versions 	map[string]*ClusterVersion
	Components  map[string]*ClusterComponent
}

type ResourceSpecKnowledge struct {
	Specs 		[]*ResourceSpec
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
	tidbType := ClusterType{"tidb", "tidb"}
	tidbV4 := ClusterVersion{"tidbV4", "v4"}

	tidbComponent := ClusterComponent{
		"tidb", "tidb",
	}

	tikvComponent := ClusterComponent{
		"tikv", "tikv",
	}

	pdComponent := ClusterComponent{
		"pd", "pd",
	}

	tidbV4Spec := ClusterVersionSpec{
		ClusterVersion: tidbV4,
		ComponentSpecs: []ClusterComponentSpec{
			{tidbComponent, ComponentConstraint{true,[]int{2}, []string{"4c8g"}, 1}},
			{tikvComponent, ComponentConstraint{true,[]int{2}, []string{"4c8g"}, 1}},
			{pdComponent, ComponentConstraint{true,[]int{2}, []string{"4c8g"}, 1}},
		},
	}

	SpecKnowledge = &ClusterSpecKnowledge {
		Specs: []*ClusterTypeSpec{{tidbType, []ClusterVersionSpec{tidbV4Spec}}},
		Types: map[string]*ClusterType{tidbType.Code:&tidbType},
		Versions: map[string]*ClusterVersion{tidbV4.Code: &tidbV4},
		Components: map[string]*ClusterComponent{tidbComponent.ComponentType:&tidbComponent,tikvComponent.ComponentType:&tikvComponent, tikvComponent.ComponentType: &tikvComponent},
	}


}
package knowledge

var SpecKnowledge *ClusterSpecKnowledge
var ParameterKnowledge *ClusterParameterKnowledge

type ClusterParameterKnowledge struct {
	Parameters []*Parameter
	Names 		map[string]*Parameter
}

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

func ParameterFromName(name string) *Parameter {
	return ParameterKnowledge.Names[name]
}

func LoadKnowledge() {
	loadSpecKnowledge()
	loadParameterKnowledge()
}

func loadSpecKnowledge () {
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
		Components: map[string]*ClusterComponent{tidbComponent.ComponentType:&tidbComponent,tikvComponent.ComponentType:&tikvComponent, pdComponent.ComponentType: &pdComponent},
	}
}

func loadParameterKnowledge() {
	ParameterKnowledge = &ClusterParameterKnowledge{
		Parameters: []*Parameter{
			{"binlog_cache_size", ParamTypeInteger, ParamUnitMb, 1024, false, []ParamValueConstraint{
				{ConstraintTypeGT, 0},
				{ConstraintTypeGT, 2048},
			}, "binlog cache size"},
			{"show_query_log", ParamTypeBoolean, ParamUnitNil, true, false, []ParamValueConstraint{}, "show_query_log"},
		},
	}

	ParameterKnowledge.Names = make(map[string]*Parameter)
	for _,v := range ParameterKnowledge.Parameters {
		ParameterKnowledge.Names[v.Name] = v
	}
}
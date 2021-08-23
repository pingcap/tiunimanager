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
	tidbv5_1_0 := ClusterVersion{"v5.1.0", "v5.1.0"}

	tidbComponent := ClusterComponent{
		"tidb", "tidb",
	}

	tikvComponent := ClusterComponent{
		"tikv", "tikv",
	}

	pdComponent := ClusterComponent{
		"pd", "pd",
	}

	tidbV5_1_0Spec := ClusterVersionSpec{
		ClusterVersion: tidbv5_1_0,
		ComponentSpecs: []ClusterComponentSpec{
			{tidbComponent, ComponentConstraint{true,[]int{2}, []string{GenSpecCode(4, 8)}, 1}},
			{tikvComponent, ComponentConstraint{true,[]int{2}, []string{GenSpecCode(4, 8)}, 1}},
			{pdComponent, ComponentConstraint{true,[]int{2}, []string{GenSpecCode(16, 32)}, 1}},
		},
	}

	dmType := ClusterType{"dm", "dm"}
	dmVx := ClusterVersion{"dmVx", "测试版本"}

	dmComponent := ClusterComponent{
		"dm", "dm",
	}

	dmVxSpec := ClusterVersionSpec{
		ClusterVersion: dmVx,
		ComponentSpecs: []ClusterComponentSpec{
			{dmComponent, ComponentConstraint{true,[]int{2}, []string{GenSpecCode(4, 8)}, 1}},
			},
	}

	SpecKnowledge = &ClusterSpecKnowledge {
		Specs: []*ClusterTypeSpec{{tidbType, []ClusterVersionSpec{tidbV5_1_0Spec}},{dmType, []ClusterVersionSpec{dmVxSpec}}},
		Types: map[string]*ClusterType{tidbType.Code:&tidbType, dmType.Code:&dmType},
		Versions: map[string]*ClusterVersion{tidbv5_1_0.Code: &tidbv5_1_0, dmVx.Code: &dmVx},
		Components: map[string]*ClusterComponent{tidbComponent.ComponentType:&tidbComponent,
			tikvComponent.ComponentType:&tikvComponent,
			pdComponent.ComponentType: &pdComponent,
			dmComponent.ComponentType: &dmComponent,
		},
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
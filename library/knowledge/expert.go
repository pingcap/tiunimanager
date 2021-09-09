package knowledge

import "github.com/pingcap-inc/tiem/library/framework"

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

func ClusterTypeSpecFromCode(code string) *ClusterTypeSpec {
	for _, v := range SpecKnowledge.Specs {
		if code == v.ClusterType.Code {
			return v
		}
	}
	framework.Log().Errorf("Unexpected code of cluster type: %s", code)
	return nil
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

func ParameterFromName(name string) *Parameter {
	return ParameterKnowledge.Names[name]
}

func LoadKnowledge() {
	loadSpecKnowledge()
	loadParameterKnowledge()
}

func loadSpecKnowledge () {
	tidbType := ClusterType{"tidb", "tidb"}
	tidbV4_0_12 := ClusterVersion{"v4.0.12", "v4.0.12"}
	tidbV5_0_0 := ClusterVersion{"v5.0.0", "v5.0.0"}

	tidbComponent := ClusterComponent{
		"TiDB", "TiDB",
	}

	tikvComponent := ClusterComponent{
		"TiKV", "TiKV",
	}

	pdComponent := ClusterComponent{
		"PD", "PD",
	}

	tiFlashComponent := ClusterComponent{
		"TiFlash", "TiFlash",
	}

	tiCdcComponent := ClusterComponent{
		"TiCDC", "TiCDC",
	}

	tidbV4_0_12_Spec := ClusterVersionSpec{
		ClusterVersion: tidbV4_0_12,
		ComponentSpecs: []ClusterComponentSpec{
			{tidbComponent, ComponentConstraint{true,[]int{3}, []string{
				GenSpecCode(4, 8),
				GenSpecCode(8, 16),
				GenSpecCode(8, 32),
				GenSpecCode(16, 32),
			}, 1}},
			{tikvComponent, ComponentConstraint{true,[]int{3}, []string{
				GenSpecCode(8, 32),
				GenSpecCode(8, 64),
				GenSpecCode(16, 128),
			}, 1}},
			{pdComponent, ComponentConstraint{true,[]int{3}, []string{
				GenSpecCode(4, 8),
				GenSpecCode(8, 16),
			}, 1}},
			{tiFlashComponent, ComponentConstraint{false,[]int{3}, []string{
				GenSpecCode(4, 32),
				GenSpecCode(8, 64),
				GenSpecCode(16, 128),
			}, 0}},
			{tiCdcComponent, ComponentConstraint{false,[]int{3}, []string{
				GenSpecCode(8, 16),
				GenSpecCode(16, 64),
			}, 0}},
		},
	}
	tidbV5_0_0_Spec := ClusterVersionSpec{
		ClusterVersion: tidbV5_0_0,
		ComponentSpecs: []ClusterComponentSpec{
			{tidbComponent, ComponentConstraint{true,[]int{3}, []string{
				GenSpecCode(4, 8),
				GenSpecCode(8, 16),
				GenSpecCode(8, 32),
				GenSpecCode(16, 32),
			}, 1}},
			{tikvComponent, ComponentConstraint{true,[]int{3}, []string{
				GenSpecCode(8, 32),
				GenSpecCode(8, 64),
				GenSpecCode(16, 128),
			}, 1}},
			{pdComponent, ComponentConstraint{true,[]int{3}, []string{
				GenSpecCode(4, 8),
				GenSpecCode(8, 16),
			}, 1}},
			{tiFlashComponent, ComponentConstraint{false,[]int{3}, []string{
				GenSpecCode(4, 32),
				GenSpecCode(8, 64),
				GenSpecCode(16, 128),
			}, 0}},
			{tiCdcComponent, ComponentConstraint{false,[]int{3}, []string{
				GenSpecCode(8, 16),
				GenSpecCode(16, 64),
			}, 0}},
		},
	}

	SpecKnowledge = &ClusterSpecKnowledge {
		Specs: []*ClusterTypeSpec{{tidbType, []ClusterVersionSpec{tidbV4_0_12_Spec, tidbV5_0_0_Spec}}},
		Types: map[string]*ClusterType{tidbType.Code:&tidbType},
		Versions: map[string]*ClusterVersion{tidbV4_0_12.Code: &tidbV4_0_12, tidbV5_0_0.Code: &tidbV5_0_0},
		Components: map[string]*ClusterComponent{tidbComponent.ComponentType:&tidbComponent,
			tikvComponent.ComponentType:&tikvComponent,
			pdComponent.ComponentType: &pdComponent,
			tiFlashComponent.ComponentType: &tiFlashComponent,
			tiCdcComponent.ComponentType: &tiCdcComponent,
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
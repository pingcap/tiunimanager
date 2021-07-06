package domain

type ClusterKnowledge struct {
	Specs 		[]*ClusterTypeSpec
	Types 		map[string]*ClusterType
	Versions 	map[string]*ClusterVersion
	Components  map[string]*ClusterComponent
}

var Knowledge *ClusterKnowledge

func (k ClusterKnowledge) ClusterTypeFromCode(code string) *ClusterType {
	return k.Types[code]
}

func (k ClusterKnowledge) ClusterVersionFromCode(code string) *ClusterVersion {
	return k.Versions[code]
}

func (k ClusterKnowledge) ClusterComponentFromCode(componentType string) *ClusterComponent {
	return k.Components[componentType]
}

func (k ClusterKnowledge) SortedTypesKnowledge() []*ClusterTypeSpec {
	slice := make([]*ClusterTypeSpec, len(k.Specs), len(k.Specs))
	for _, v := range k.Specs  {
		slice = append(slice, v)
	}

	return slice
}

type ClusterTypeSpec struct {
	ClusterType *ClusterType
	VersionSpec []*ClusterVersionSpec
}

type ClusterVersionSpec struct {
	ClusterVersion *ClusterVersion
	Components []*ClusterComponentSpec
}

type ClusterComponentSpec struct {
	ClusterComponent *ClusterComponent
	ComponentConstraint *ComponentConstraint
}

type ComponentConstraint struct {
	ComponentRequired 		bool
	SuggestedNodeQuantities []int
	AvailableSpecCodes		[]string
	MinZoneQuantity			int
}

func loadKnowledge () {
	Knowledge = &ClusterKnowledge{
		// TODO
	}
}

func MeetTheSpec (cluster *Cluster) error {
	return nil
}

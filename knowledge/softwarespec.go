package knowledge

type ClusterType struct {
	Code string
	Name string
}

type ClusterVersion struct {
	Code string
	Name string
}

type ClusterComponent struct {
	ComponentType 			string
	ComponentName			string
}

type ClusterTypeSpec struct {
	ClusterType  ClusterType
	VersionSpecs []ClusterVersionSpec
}

type ClusterVersionSpec struct {
	ClusterVersion ClusterVersion
	ComponentSpecs []ClusterComponentSpec
}

type ClusterComponentSpec struct {
	ClusterComponent    ClusterComponent
	ComponentConstraint ComponentConstraint
}

type ComponentConstraint struct {
	ComponentRequired 		bool
	SuggestedNodeQuantities []int
	AvailableSpecCodes		[]string
	MinZoneQuantity			int
}

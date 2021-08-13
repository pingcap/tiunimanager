package knowledge

type ClusterType struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ClusterVersion struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ClusterComponent struct {
	ComponentType 			string `json:"componentType"`
	ComponentName			string `json:"componentName"`
}

type ClusterTypeSpec struct {
	ClusterType  ClusterType `json:"clusterType"`
	VersionSpecs []ClusterVersionSpec`json:"versionSpecs"`
}

type ClusterVersionSpec struct {
	ClusterVersion ClusterVersion  `json:"clusterVersion"`
	ComponentSpecs []ClusterComponentSpec `json:"componentSpecs"`
}

type ClusterComponentSpec struct {
	ClusterComponent    ClusterComponent  `json:"clusterComponent"`
	ComponentConstraint ComponentConstraint  `json:"componentConstraint"`
}

type ComponentConstraint struct {
	ComponentRequired 		bool  `json:"componentRequired"`
	SuggestedNodeQuantities []int  `json:"suggestedNodeQuantities"`
	AvailableSpecCodes		[]string  `json:"availableSpecCodes"`
	MinZoneQuantity			int  `json:"minZoneQuantity"`
}

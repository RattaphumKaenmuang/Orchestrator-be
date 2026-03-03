package store

import "orchestrator/internal/model"

var groups map[string]*model.Group

func Init() {
	groups = make(map[string]*model.Group)
}

func ensureInit() {
	if groups == nil {
		Init()
	}
}

func CreateGroup(name string) *model.Group {
	ensureInit()

	g := &model.Group{
		Name:      name,
		Instances: []model.Instance{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) (*model.Group, bool) {
	ensureInit()

	g, ok := groups[name]
	return g, ok
}

func ListGroups() []*model.Group {
	ensureInit()

	out := make([]*model.Group, 0, len(groups))
	for _, g := range groups {
		out = append(out, g)
	}
	return out
}

func AddInstance(groupName string, instance model.Instance) {
	ensureInit()
	groups[groupName].Instances = append(groups[groupName].Instances, instance)
}

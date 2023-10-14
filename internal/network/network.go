package network

import "test/model"

type Network interface {
	FindDomainIPs(dom model.Domain) []string
	CreateInterface(driver string) (*model.Interface,error)
}
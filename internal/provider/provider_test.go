package provider_test

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	minecraftprovider "github.com/hashicraft/terraform-provider-minecraft/internal/provider"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"minecraft": providerserver.NewProtocol6WithError(minecraftprovider.New("test")()),
}

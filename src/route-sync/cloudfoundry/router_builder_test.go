package cloudfoundry

import (
	"errors"
	"route-sync/cloudfoundry/tcp"
	tcpfakes "route-sync/cloudfoundry/tcp/fakes"
	"route-sync/config"
	cfConfig "code.cloudfoundry.org/route-registrar/config"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/route-registrar/messagebus"
	messagebusfakes "code.cloudfoundry.org/route-registrar/messagebus/fakes"
	uaa "code.cloudfoundry.org/uaa-go-client"
	uaaconfig "code.cloudfoundry.org/uaa-go-client/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("CloudFoundryRouterBuilder", func() {
	var (
		logger lager.Logger
		cfg    = &config.Config{
			RawNatsServers:            "[{\"Host\": \"host\",\"User\": \"user\", \"Password\": \"password\"}]",
			NatsServers:               []cfConfig.MessageBusServer{{Host: "host", User: "user", Password: "password"}},
			RoutingApiUrl:             "https://api.cf.example.org",
			CloudFoundryAppDomainName: "apps.cf.example.org",
			UAAApiURL:                 "https://uaa.cf.example.org",
			RoutingAPIUsername:        "routeUser",
			RoutingAPIClientSecret:    "aabbcc",
			SkipTLSVerification:       true,
			KubeConfigPath:            "~/.config/kube",
		}
	)
	BeforeEach(func() {
		logger = lagertest.NewTestLogger("")
	})
	Context("TCPRouter", func() {
		var (
			uaaClientBuilderFunc func(logger lager.Logger, cfg *uaaconfig.Config, clock clock.Clock) (uaa.Client, error)
			tcpRouterBuilderFunc func(uaaClient uaa.Client, routingApiUrl string, skipTlsVerification bool) (tcp.Router, error)
			fakeRouter           tcp.Router
		)
		BeforeEach(func() {
			fakeRouter = &tcpfakes.FakeRouter{}
			uaaClientBuilderFunc = func(logger lager.Logger, cfg *uaaconfig.Config, clock clock.Clock) (uaa.Client, error) {
				return nil, nil
			}
			tcpRouterBuilderFunc = func(uaaClient uaa.Client, routingApiUrl string, skipTlsVerification bool) (tcp.Router, error) {
				return fakeRouter, nil
			}
		})
		It("returns a TCP router", func() {
			routingBuilder := NewCloudFoundryRoutingBuilder(cfg, logger)
			client := routingBuilder.CreateTCPRouter(uaaClientBuilderFunc, tcpRouterBuilderFunc)
			Expect(client).To(Equal(fakeRouter))
		})
		It("panics when uaa client fails", func() {
			uaaClientBuilderFunc = func(logger lager.Logger, cfg *uaaconfig.Config, clock clock.Clock) (uaa.Client, error) {
				return nil, errors.New("")
			}
			routingBuilder := NewCloudFoundryRoutingBuilder(cfg, logger)
			defer func() {
				recover()
				Eventually(logger).Should(gbytes.Say("creating UAA client"))
			}()
			routingBuilder.CreateTCPRouter(uaaClientBuilderFunc, tcpRouterBuilderFunc)
		})
		It("panics when tcp router creation fails", func() {
			tcpRouterBuilderFunc = func(uaaClient uaa.Client, routingApiUrl string, skipTlsVerification bool) (tcp.Router, error) {
				return nil, errors.New("")
			}
			routingBuilder := NewCloudFoundryRoutingBuilder(cfg, logger)
			defer func() {
				recover()
				Eventually(logger).Should(gbytes.Say("creating TCP router"))
			}()
			routingBuilder.CreateTCPRouter(uaaClientBuilderFunc, tcpRouterBuilderFunc)
		})
	})

	Context("HTTPRouter", func() {
		var (
			fakeMessageBus        *messagebusfakes.FakeMessageBus
			messageBusBuilderFunc func(logger lager.Logger) messagebus.MessageBus
		)

		BeforeEach(func() {
			fakeMessageBus = &messagebusfakes.FakeMessageBus{}
			messageBusBuilderFunc = func(logger lager.Logger) messagebus.MessageBus {
				return fakeMessageBus
			}
		})

		It("returns correct HTTP router", func() {
			routingBuilder := NewCloudFoundryRoutingBuilder(cfg, logger)
			mb := routingBuilder.CreateHTTPRouter(messageBusBuilderFunc)
			Expect(mb).To(Equal(fakeMessageBus))
		})
	})
})

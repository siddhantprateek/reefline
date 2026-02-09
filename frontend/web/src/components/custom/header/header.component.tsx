import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import {
  NavigationMenu,
  NavigationMenuContent,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  NavigationMenuTrigger,
} from '@/components/ui/navigation-menu';
import { Menu, X, ChevronDown, Ship, Anchor, Package, Globe, LayoutTemplate, MessageCircleQuestion, LifeBuoy, Landmark, ArrowUpRight } from 'lucide-react';

const services = [
  { title: 'Ocean Freight', description: 'Global container shipping solutions', icon: Ship },
  { title: 'Port Services', description: 'Efficient port handling and storage', icon: Anchor },
  { title: 'Cargo Tracking', description: 'Real-time shipment monitoring', icon: Package },
  { title: 'Global Network', description: 'Worldwide logistics coverage', icon: Globe },
];

const resources = [
  { title: 'Brand', description: 'Assets, examples and guides', icon: LayoutTemplate },
  { title: 'FAQ', description: 'Answers to common questions', icon: MessageCircleQuestion },
  { title: 'Help & Support', description: 'Guides, articles and more', icon: LifeBuoy },
  { title: 'Governance', description: 'The Aave Governance forum', icon: Landmark },
];

export function Header() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  return (
    <header className="fixed top-0 left-0 right-0 z-50 px-4 md:px-8 pt-6">
      <div className="mx-auto max-w-7xl flex items-center justify-between">

        {/* Logo */}
        <Link to="/" className="flex items-center gap-2 group">
          <div className="relative">
            <span className="text-xl md:text-2xl font-extrabold text-white tracking-tight flex items-center gap-2">
              Reefline
            </span>
          </div>
        </Link>

        {/* Desktop Navigation - Centered Glass Pill */}
        <nav className="hidden lg:flex items-center absolute left-1/2 top-6 -translate-x-1/2 z-50">
          <div className="glass rounded-md px-1 py-1 border border-white/10 shadow-lg shadow-black/5 backdrop-blur-md bg-white/5">
            <NavigationMenu>
              <NavigationMenuList className="gap-1">
                <NavigationMenuItem>
                  <Link to="/">
                    <NavigationMenuLink className="px-5 py-2 text-sm font-medium text-white/90 hover:text-white hover:bg-white/10 rounded-md transition-all duration-200">
                      Home
                    </NavigationMenuLink>
                  </Link>
                </NavigationMenuItem>

                <NavigationMenuItem>
                  <NavigationMenuTrigger className="px-4 py-2 text-sm font-medium text-white/90 bg-transparent hover:!bg-white/10 hover:text-white focus:!bg-white/10 focus:text-white data-[state=open]:!bg-white/10 data-[state=open]:text-white rounded-md transition-colors">
                    Services
                  </NavigationMenuTrigger>
                  <NavigationMenuContent>
                    <div className="flex w-[600px] rounded-2xl mt-2 overflow-hidden p-0">
                      {/* Left Column: Services List */}
                      <div className="flex-1 p-4 space-y-2">
                        {services.map((service) => (
                          <Link
                            key={service.title}
                            to="#"
                            className="group flex items-start gap-3 p-3 rounded-xl hover:bg-gray-100 transition-all duration-200"
                          >
                            <div className="p-2 rounded-lg bg-orange-50 text-orange-600 group-hover:bg-orange-100 group-hover:text-orange-700 transition-all duration-200">
                              <service.icon className="w-5 h-5" />
                            </div>
                            <div>
                              <h4 className="text-sm font-semibold text-gray-900">{service.title}</h4>
                              <p className="text-xs text-gray-500 mt-0.5">{service.description}</p>
                            </div>
                          </Link>
                        ))}
                      </div>

                      {/* Right Column: Banner Area */}
                      <div className="w-[240px] relative bg-gradient-to-br from-cyan-600 via-blue-600 to-indigo-600 p-6 flex flex-col justify-end overflow-hidden">
                        {/* Decorative Elements */}
                        <div className="absolute top-0 right-0 w-32 h-32 bg-white/10 rounded-md blur-2xl -translate-y-1/2 translate-x-1/2"></div>
                        <div className="absolute bottom-10 left-0 w-24 h-24 bg-cyan-400/20 rounded-md blur-2xl translate-y-1/2 -translate-x-1/2"></div>
                        <div className="absolute inset-0 bg-white/10 pattern-grid-lg opacity-20"></div>

                        <div className="relative z-10 mt-auto">
                          <h4 className="text-lg font-bold text-white mb-2">Logistics</h4>
                          <p className="text-xs text-white/80 leading-relaxed">
                            End-to-end shipping and port services for global trade.
                          </p>
                        </div>
                      </div>
                    </div>
                  </NavigationMenuContent>
                </NavigationMenuItem>

                <NavigationMenuItem>
                  <NavigationMenuTrigger className="px-4 py-2 text-sm font-medium text-white/90 bg-transparent hover:!bg-white/10 hover:text-white focus:!bg-white/10 focus:text-white data-[state=open]:!bg-white/10 data-[state=open]:text-white rounded-md transition-colors">
                    Resources
                  </NavigationMenuTrigger>
                  <NavigationMenuContent>
                    <div className="flex w-[600px] rounded-2xl mt-2 overflow-hidden p-0">
                      {/* Left Column: Resources List */}
                      <div className="flex-1 p-4 space-y-2">
                        {resources.map((resource) => (
                          <Link
                            key={resource.title}
                            to="#"
                            className="group flex items-start gap-3 p-3 rounded-xl hover:bg-gray-100 transition-all duration-200"
                          >
                            <div className="p-2 rounded-lg bg-gray-100 text-gray-600 group-hover:bg-purple-100 group-hover:text-purple-600 transition-all duration-200">
                              <resource.icon className="w-5 h-5" />
                            </div>
                            <div>
                              <h4 className="text-sm font-semibold text-gray-900">{resource.title}</h4>
                              <p className="text-xs text-gray-500 mt-0.5">{resource.description}</p>
                            </div>
                          </Link>
                        ))}
                      </div>

                      {/* Right Column: Banner Area */}
                      <div className="w-[240px] relative bg-gradient-to-br from-indigo-600 via-purple-600 to-blue-600 p-6 flex flex-col justify-end overflow-hidden">
                        {/* Decorative Elements */}
                        <div className="absolute top-0 right-0 w-32 h-32 bg-purple-500/40 rounded-md blur-2xl -translate-y-1/2 translate-x-1/2"></div>
                        <div className="absolute bottom-10 left-0 w-24 h-24 bg-indigo-500/40 rounded-md blur-2xl translate-y-1/2 -translate-x-1/2"></div>
                        <div className="absolute inset-0 bg-white/10 pattern-grid-lg opacity-20"></div>

                        <div className="relative z-10 mt-auto">
                          <h4 className="text-lg font-bold text-white mb-2">Platform</h4>
                          <p className="text-xs text-white/80 leading-relaxed">
                            Explore our comprehensive ecosystem of tools and services.
                          </p>
                        </div>
                      </div>
                    </div>
                  </NavigationMenuContent>
                </NavigationMenuItem>

                <NavigationMenuItem>
                  <Link to="#">
                    <NavigationMenuLink className="px-5 py-2 text-sm font-medium text-white/90 hover:text-white hover:bg-white/10 rounded-md transition-all duration-200">
                      About
                    </NavigationMenuLink>
                  </Link>
                </NavigationMenuItem>

                <NavigationMenuItem>
                  <Link to="#">
                    <NavigationMenuLink className="px-5 py-2 text-sm font-medium text-white/90 hover:text-white hover:bg-white/10 rounded-md transition-all duration-200">
                      Career
                    </NavigationMenuLink>
                  </Link>
                </NavigationMenuItem>
              </NavigationMenuList>
            </NavigationMenu>
          </div>
        </nav>

        {/* CTA Button & Mobile Toggle */}
        <div className="flex items-center gap-4">
          <Link to="#" className="hidden lg:flex group items-center">
            <div className="bg-orange-500 text-white font-semibold px-5 py-2.5 rounded-lg hover:bg-orange-600 transition-colors duration-300">
              Get Started
            </div>
            <div className="bg-orange-500 text-white p-2.5 rounded-lg group-hover:bg-orange-600 transition-colors duration-300">
              <ArrowUpRight className="w-6 h-6 transition-transform duration-300 group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
            </div>
          </Link>

          <button
            className="lg:hidden p-2 rounded-lg text-white hover:bg-white/10 transition-colors"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          >
            {mobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
          </button>
        </div>

        {/* Mobile Menu Dropdown */}
        {mobileMenuOpen && (
          <div className="absolute left-4 right-4 top-[calc(100%+1rem)] glass-dark rounded-2xl p-4 animate-fade-in-up border border-white/10 z-50">
            <nav className="space-y-2">
              <Link
                to="/"
                className="block px-4 py-3 text-white font-medium hover:bg-white/10 rounded-xl transition-colors"
                onClick={() => setMobileMenuOpen(false)}
              >
                Home
              </Link>

              <div className="px-4 py-3">
                <div className="flex items-center justify-between text-white font-medium">
                  <span>Services</span>
                  <ChevronDown className="w-4 h-4" />
                </div>
                <div className="mt-2 space-y-1 pl-4">
                  {services.map((service) => (
                    <Link
                      key={service.title}
                      to="#"
                      className="block py-2 text-white/70 hover:text-white transition-colors"
                      onClick={() => setMobileMenuOpen(false)}
                    >
                      {service.title}
                    </Link>
                  ))}
                </div>
              </div>

              <div className="px-4 py-3">
                <div className="flex items-center justify-between text-white font-medium">
                  <span>Resources</span>
                  <ChevronDown className="w-4 h-4" />
                </div>
                <div className="mt-2 space-y-1 pl-4">
                  {resources.map((resource) => (
                    <Link
                      key={resource.title}
                      to="#"
                      className="block py-2 text-white/70 hover:text-white transition-colors"
                      onClick={() => setMobileMenuOpen(false)}
                    >
                      {resource.title}
                    </Link>
                  ))}
                </div>
              </div>

              <Link
                to="#"
                className="block px-4 py-3 text-white font-medium hover:bg-white/10 rounded-xl transition-colors"
                onClick={() => setMobileMenuOpen(false)}
              >
                About
              </Link>

              <Link
                to="#"
                className="block px-4 py-3 text-white font-medium hover:bg-white/10 rounded-xl transition-colors"
                onClick={() => setMobileMenuOpen(false)}
              >
                Career
              </Link>

              <div className="pt-4 border-t border-white/10">
                <Button className="w-full btn-orange-gradient text-white font-semibold py-3 rounded-xl">
                  Get Started
                </Button>
              </div>
            </nav>
          </div>
        )}
      </div>
    </header>
  );
}

export default Header;

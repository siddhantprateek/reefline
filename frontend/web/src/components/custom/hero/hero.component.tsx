import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ArrowUpRight, MapPin, Calendar, Container, Search, ArrowRightLeft, Anchor, Globe } from 'lucide-react';

// Partner logos as text (in production, these would be actual logos)
const partners = [
  { name: 'Ferrari', style: 'font-serif italic font-bold' },
  { name: 'TOYOTA', style: 'font-bold tracking-wider' },
  { name: 'TESLA', style: 'font-bold tracking-[0.3em]' },
  { name: 'HIGER', style: 'font-bold italic' },
  { name: 'Marcopolo', style: 'font-bold' },
];

export function Hero() {
  const [activeTab, setActiveTab] = useState<'tracking' | 'schedules'>('schedules');

  return (
    <div>
      {/* Hero Section Container with Padding/Margins */}
      <section className="relative p-2 md:px-4 md:pt-1">
        <div className="relative rounded-xl overflow-hidden min-h-[85vh] ring-1 ring-black/5">
          {/* Background Image with Overlay */}
          <div className="absolute inset-0 z-0">
            <div
              className="absolute inset-0 bg-cover bg-center bg-no-repeat transition-transform duration-10000 motion-safe:animate-ken-burns"
              style={{
                backgroundImage: `url('https://images.unsplash.com/photo-1494412574643-ff11b0a5c1c3?q=80&w=2070&auto=format&fit=crop')`,
              }}
            />
            <div className="absolute inset-0 hero-gradient opacity-90" />
          </div>

          {/* Content */}
          <div className="relative z-10 mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 pt-32 pb-20">
            <div className="grid lg:grid-cols-2 gap-12 items-center min-h-[60vh]">
              {/* Left Column - Text Content */}
              <div className="space-y-8 animate-fade-in-up">
                {/* Badge */}
                <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full glass text-white/90 text-sm font-medium border border-white/10 backdrop-blur-md">
                  <Globe className="w-4 h-4 text-orange-400" />
                  <span>Unmatched Worldwide Reach</span>
                </div>

                {/* Heading */}
                <h1 className="text-5xl md:text-6xl font-light leading-[1.05] mt-40 tracking-tight">
                  <span className="text-white drop-shadow-sm">Global ocean cargo</span>
                  <br />
                  <span className="text-white/90">— Efficient, on time</span>
                  <br />
                  <span className="text-orange-100/90">and trusted</span>
                </h1>

                {/* Description */}
                <p className="text-lg text-white/80 max-w-lg leading-relaxed font-light">
                  Navigate the complexities of international shipping with our reliable ocean cargo services.
                  Real-time tracking, competitive rates, and seamless delivery worldwide.
                </p>

                {/* Stats */}
                <div className="flex flex-wrap gap-12 pt-6 border-t border-white/10">
                  <div className="space-y-1">
                    <div className="text-3xl font-bold text-white">150+</div>
                    <div className="text-sm text-white/60 font-medium tracking-wide">Ports Covered</div>
                  </div>
                  <div className="space-y-1">
                    <div className="text-3xl font-bold text-white">98%</div>
                    <div className="text-sm text-white/60 font-medium tracking-wide">On-Time Delivery</div>
                  </div>
                  <div className="space-y-1">
                    <div className="text-3xl font-bold text-white">24/7</div>
                    <div className="text-sm text-white/60 font-medium tracking-wide">Support Available</div>
                  </div>
                </div>
              </div>

              {/* Right Column - Search Card */}
              <div className="lg:justify-self-end w-full max-w-md animate-slide-in-right" style={{ animationDelay: '0.2s' }}>
                <div className="glass rounded-[2rem] p-6 shadow-2xl shadow-black/20 border border-white/10 backdrop-blur-xl bg-black/10">
                  {/* Tabs */}
                  <div className="flex mb-6 p-1 bg-black/20 rounded-full border border-white/5">
                    <button
                      onClick={() => setActiveTab('tracking')}
                      className={`flex-1 py-3 px-4 rounded-full text-sm font-medium transition-all duration-300 ${activeTab === 'tracking'
                        ? 'bg-orange-500 text-white shadow-lg'
                        : 'text-white/70 hover:text-white hover:bg-white/5'
                        }`}
                    >
                      Tracking
                    </button>
                    <button
                      onClick={() => setActiveTab('schedules')}
                      className={`flex-1 py-3 px-4 rounded-full text-sm font-medium transition-all duration-300 ${activeTab === 'schedules'
                        ? 'bg-orange-500 text-white shadow-lg'
                        : 'text-white/70 hover:text-white hover:bg-white/5'
                        }`}
                    >
                      Schedules
                    </button>
                  </div>

                  {/* Form Fields */}
                  <div className="space-y-4">
                    {/* Origin */}
                    <div className="relative group">
                      <div className="absolute left-4 top-1/2 -translate-y-1/2 text-white/50 group-focus-within:text-orange-400 transition-colors">
                        <MapPin className="w-5 h-5" />
                      </div>
                      <Input
                        placeholder="Boston, United States (BDCGP)"
                        className="w-full pl-12 pr-12 py-6 bg-white/5 border-white/10 text-white placeholder:text-white/40 rounded-2xl focus:border-orange-500/50 focus:ring-orange-500/20 focus:bg-white/10 transition-all font-light"
                      />
                      <button className="absolute right-3 top-1/2 -translate-y-1/2 p-2 rounded-xl bg-white/5 text-white/70 hover:bg-orange-500 hover:text-white transition-all duration-300">
                        <ArrowRightLeft className="w-4 h-4" />
                      </button>
                    </div>

                    {/* Destination */}
                    <div className="relative group">
                      <div className="absolute left-4 top-1/2 -translate-y-1/2 text-white/50 group-focus-within:text-orange-400 transition-colors">
                        <Anchor className="w-5 h-5" />
                      </div>
                      <Input
                        placeholder="Singapore, Singapore (SGSIN)"
                        className="w-full pl-12 py-6 bg-white/5 border-white/10 text-white placeholder:text-white/40 rounded-2xl focus:border-orange-500/50 focus:ring-orange-500/20 focus:bg-white/10 transition-all font-light"
                      />
                    </div>

                    {/* Date */}
                    <div className="relative group">
                      <div className="absolute left-4 top-1/2 -translate-y-1/2 text-white/50 group-focus-within:text-orange-400 transition-colors">
                        <Calendar className="w-5 h-5" />
                      </div>
                      <Input
                        placeholder="29 Aug, 2025"
                        className="w-full pl-12 py-6 bg-white/5 border-white/10 text-white placeholder:text-white/40 rounded-2xl focus:border-orange-500/50 focus:ring-orange-500/20 focus:bg-white/10 transition-all font-light"
                      />
                    </div>

                    {/* Search Button */}
                    <Button className="w-full btn-orange-gradient text-white font-semibold py-6 rounded-2xl flex items-center justify-center gap-2 group mt-4 shadow-lg shadow-orange-500/20 hover:shadow-orange-500/40">
                      <Search className="w-5 h-5" />
                      <span className="text-base">Search Content</span>
                      <ArrowUpRight className="w-5 h-5 transition-transform duration-300 group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Partners Section */}
      <div className="relative z-10 py-12">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <p className="text-center text-sm text-gray-500 mb-6">Partners of world leading companies</p>
          <div className="flex flex-wrap items-center justify-center gap-8 md:gap-16">
            {partners.map((partner) => (
              <div
                key={partner.name}
                className={`text-xl md:text-2xl text-gray-400 partner-logo cursor-pointer ${partner.style}`}
              >
                {partner.name}
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="relative z-10 py-12 md:py-20">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="grid lg:grid-cols-2 gap-16 items-center">
            {/* Left - Text Content */}
            <div className="space-y-6">
              {/* Badge */}
              <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-orange-100 text-orange-600 text-sm font-medium">
                <Container className="w-4 h-4" />
                Service Overview
              </div>

              <h2 className="text-4xl md:text-5xl font-bold text-gray-900 leading-tight">
                Navigate global trade with trusted ocean logistics
              </h2>

              <p className="text-lg text-gray-600 leading-relaxed">
                Need to optimize production or deliver time-critical goods? Ocean Contract ensures a smoother
                supply chain with flexible setup, clear insights, and reliable global delivery.
              </p>

              <p className="text-gray-600">
                Ocean Contract provides you with access to real-time data on all your ocean lanes with its
                Allocation Portal.
              </p>

              <Button className="btn-orange-gradient text-white font-semibold px-8 py-3 rounded-full flex items-center gap-2 group mt-4">
                Ship now
                <ArrowUpRight className="w-4 h-4 transition-transform duration-300 group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
              </Button>
            </div>

            {/* Right - Image Grid */}
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-4">
                <div className="rounded-2xl overflow-hidden shadow-xl transform hover:scale-[1.02] transition-transform duration-300">
                  <img
                    src="https://images.unsplash.com/photo-1578575437130-527eed3abbec?q=80&w=400&auto=format&fit=crop"
                    alt="Container ship at sea"
                    className="w-full h-48 object-cover"
                  />
                </div>
                <div className="rounded-2xl overflow-hidden shadow-xl transform hover:scale-[1.02] transition-transform duration-300">
                  <img
                    src="https://images.unsplash.com/photo-1605732562742-3023a888e56e?q=80&w=400&auto=format&fit=crop"
                    alt="Cargo ship with containers"
                    className="w-full h-64 object-cover"
                  />
                </div>
              </div>
              <div className="pt-8 space-y-4">
                <div className="rounded-2xl overflow-hidden shadow-xl transform hover:scale-[1.02] transition-transform duration-300">
                  <img
                    src="https://images.unsplash.com/photo-1559083991-9bdef50986d5?q=80&w=400&auto=format&fit=crop"
                    alt="Aerial view of container port"
                    className="w-full h-64 object-cover"
                  />
                </div>
                <div className="rounded-2xl overflow-hidden shadow-xl transform hover:scale-[1.02] transition-transform duration-300">
                  <img
                    src="https://images.unsplash.com/photo-1601584115197-04ecc0da31d7?q=80&w=400&auto=format&fit=crop"
                    alt="Shipping containers"
                    className="w-full h-48 object-cover"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default Hero;

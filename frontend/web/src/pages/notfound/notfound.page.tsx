import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Ship, Home, ArrowLeft, Anchor, Compass, Waves } from 'lucide-react';

export function NotFoundPage() {
  return (
    <div className="min-h-screen relative overflow-hidden">
      {/* Background */}
      <div className="absolute inset-0 z-0">
        <div
          className="absolute inset-0 bg-cover bg-center bg-no-repeat"
          style={{
            backgroundImage: `url('https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?q=80&w=2070&auto=format&fit=crop')`,
          }}
        />
        <div className="absolute inset-0 hero-gradient" />
      </div>

      {/* Floating Elements */}
      <div className="absolute inset-0 z-10 overflow-hidden pointer-events-none">
        <Anchor className="absolute top-1/4 left-[10%] w-16 h-16 text-white/10 animate-float" />
        <Compass className="absolute top-1/3 right-[15%] w-20 h-20 text-white/10 animate-float" style={{ animationDelay: '1s' }} />
        <Waves className="absolute bottom-1/4 left-[20%] w-24 h-24 text-white/10 animate-float" style={{ animationDelay: '2s' }} />
        <Ship className="absolute bottom-1/3 right-[10%] w-16 h-16 text-white/10 animate-float" style={{ animationDelay: '0.5s' }} />
      </div>

      {/* Content */}
      <div className="relative z-20 min-h-screen flex flex-col items-center justify-center px-4">
        {/* Logo */}
        <Link to="/" className="flex items-center gap-2 mb-12 group">
          <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-orange-500 to-orange-600 flex items-center justify-center shadow-lg shadow-orange-500/30 group-hover:shadow-orange-500/50 transition-all duration-300">
            <Ship className="w-6 h-6 text-white" />
          </div>
          <span className="text-2xl font-bold text-white tracking-tight">Reefline</span>
        </Link>

        {/* 404 Text */}
        <div className="text-center space-y-6 animate-fade-in-up">
          <div className="relative">
            <h1 className="text-[150px] md:text-[200px] font-bold text-white/10 leading-none select-none">
              404
            </h1>
            <div className="absolute inset-0 flex items-center justify-center">
              <div className="text-center">
                <div className="w-24 h-24 mx-auto mb-4 rounded-full bg-orange-500/20 flex items-center justify-center animate-pulse">
                  <Anchor className="w-12 h-12 text-orange-400" />
                </div>
              </div>
            </div>
          </div>

          <h2 className="text-3xl md:text-4xl font-bold text-white">
            Lost at Sea?
          </h2>

          <p className="text-lg text-white/70 max-w-md mx-auto">
            The page you're looking for seems to have drifted away.
            Let's navigate you back to safe harbor.
          </p>

          {/* Buttons */}
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-8">
            <Button
              asChild
              className="btn-orange-gradient text-white font-semibold px-8 py-4 rounded-full flex items-center gap-2 group"
            >
              <Link to="/">
                <Home className="w-5 h-5" />
                Back to Home
              </Link>
            </Button>

            <Button
              asChild
              variant="outline"
              className="bg-white/10 border-white/20 text-white hover:bg-white/20 font-semibold px-8 py-4 rounded-full flex items-center gap-2"
            >
              <Link to="/">
                <ArrowLeft className="w-5 h-5" />
                Go Back
              </Link>
            </Button>
          </div>
        </div>

        {/* Quick Links */}
        <div className="mt-16 glass rounded-2xl p-6 animate-fade-in-up" style={{ animationDelay: '0.3s' }}>
          <p className="text-white/60 text-sm mb-4 text-center">Quick Navigation</p>
          <div className="flex flex-wrap items-center justify-center gap-4">
            <Link
              to="/"
              className="px-4 py-2 rounded-full bg-white/10 text-white/80 hover:bg-white/20 hover:text-white text-sm transition-all duration-200"
            >
              Home
            </Link>
            <Link
              to="/"
              className="px-4 py-2 rounded-full bg-white/10 text-white/80 hover:bg-white/20 hover:text-white text-sm transition-all duration-200"
            >
              Services
            </Link>
            <Link
              to="/"
              className="px-4 py-2 rounded-full bg-white/10 text-white/80 hover:bg-white/20 hover:text-white text-sm transition-all duration-200"
            >
              Track Shipment
            </Link>
            <Link
              to="/"
              className="px-4 py-2 rounded-full bg-white/10 text-white/80 hover:bg-white/20 hover:text-white text-sm transition-all duration-200"
            >
              Contact Us
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}

export default NotFoundPage;

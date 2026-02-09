import { Header, Hero, Footer } from '@/components';

export function HomePage() {
  return (
    <div className="min-h-screen">
      <Header />
      <main>
        <Hero />
      </main>
      <Footer />
    </div>
  );
}

export default HomePage;

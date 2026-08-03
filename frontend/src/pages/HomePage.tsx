import { HeroSection } from '../components/design/HeroSection';
import { TickerSection } from '../components/design/TickerSection';
import { FeaturesSection } from '../components/design/FeaturesSection';
import { StatsSection } from '../components/design/StatsSection';

export default function HomePage() {
  return (
    <>
      <HeroSection />
      <TickerSection />
      <FeaturesSection />
      <StatsSection />
    </>
  );
}

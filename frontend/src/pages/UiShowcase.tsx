import { Crown, Zap, Users, Check, ArrowRight } from 'lucide-react';
import { Button } from '../components/ui/button';
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from '../components/ui/card';

const plans = [
  {
    icon: <Zap className="h-5 w-5" />,
    name: 'Starter',
    desc: 'For casual gamers',
    price: 'Free',
    per: '',
    features: ['Basic optimization', 'RAM cleanup', 'Startup manager'],
    cta: 'Download',
    variant: 'secondary' as const,
  },
  {
    icon: <Crown className="h-5 w-5" />,
    name: 'Lifetime Pro',
    desc: 'One-time payment, forever',
    price: '$15',
    per: ' once',
    features: ['Everything in Starter', 'AI tuning', 'FPS-boost guarantee', 'Priority support'],
    cta: 'Get Lifetime Pro',
    variant: 'default' as const,
  },
  {
    icon: <Users className="h-5 w-5" />,
    name: 'Team',
    desc: 'For esports teams & cafés',
    price: '$149.99',
    per: '',
    features: ['5 PC licenses', 'Team dashboard', 'API access', 'Dedicated manager'],
    cta: 'Buy Team',
    variant: 'outline' as const,
  },
];

export default function UiShowcase() {
  return (
    <section className="tab-content active">
      <div className="max-w-5xl mx-auto px-6 py-16 space-y-14">
        <header className="space-y-2">
          <h1 className="text-3xl font-bold text-white">shadcn/ui + Tailwind — Aura theme</h1>
          <p className="text-gray-400">
            Reference page. These live in{' '}
            <code className="text-aura-400">src/components/ui</code> and are themed to your
            dark/emerald brand. Your existing pages are untouched.
          </p>
        </header>

        <div className="space-y-4">
          <h2 className="text-xs uppercase tracking-wider text-gray-500">Buttons</h2>
          <div className="flex flex-wrap gap-3">
            <Button>
              Get Lifetime Pro <ArrowRight className="h-4 w-4" />
            </Button>
            <Button variant="secondary">Secondary</Button>
            <Button variant="outline">Outline</Button>
            <Button variant="ghost">Ghost</Button>
            <Button variant="discord">Continue with Discord</Button>
            <Button variant="destructive">Destructive</Button>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <Button size="sm">Small</Button>
            <Button>Default</Button>
            <Button size="lg">Large</Button>
            <Button disabled>Disabled</Button>
          </div>
        </div>

        <div className="space-y-4">
          <h2 className="text-xs uppercase tracking-wider text-gray-500">Cards</h2>
          <div className="grid gap-5 sm:grid-cols-3">
            {plans.map((p) => (
              <Card key={p.name}>
                <CardHeader>
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-aura-500/10 text-aura-400">
                    {p.icon}
                  </div>
                  <CardTitle>{p.name}</CardTitle>
                  <CardDescription>{p.desc}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="text-3xl font-bold text-white">
                    {p.price}
                    <span className="text-sm font-normal text-gray-500">{p.per}</span>
                  </div>
                  <ul className="space-y-2 text-sm text-gray-400">
                    {p.features.map((f) => (
                      <li key={f} className="flex gap-2">
                        <Check className="h-4 w-4 shrink-0 text-aura-400" />
                        {f}
                      </li>
                    ))}
                  </ul>
                </CardContent>
                <CardFooter>
                  <Button variant={p.variant} className="w-full">
                    {p.cta}
                  </Button>
                </CardFooter>
              </Card>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

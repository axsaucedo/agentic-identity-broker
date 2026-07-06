import Heading from '@theme/Heading';
import Link from '@docusaurus/Link';
import styles from './styles.module.css';

const capabilityCards = [
  {
    title: 'Delegated access, governed',
    description:
      'A user grants an agent access to specific third-party services at a chosen permission set, with optional expiry. One grant per user and agent, revocable at any time.',
  },
  {
    title: 'An encrypted token vault',
    description:
      'The broker stores third-party access and refresh tokens encrypted at rest with envelope encryption bound to each service, so agents never hold long-lived provider credentials.',
  },
  {
    title: 'Token exchange at the edge',
    description:
      'A gateway swaps an agent token for the right third-party token per request using RFC 8693, enforced by policy — transparently, with an optional Envoy sidecar.',
  },
];

const pathCards = [
  {
    title: 'Introduction',
    href: '/docs/introduction',
    description:
      'What the broker is, the delegation problem it solves, and when to reach for it.',
  },
  {
    title: 'Concepts',
    href: '/docs/concepts',
    description:
      'The delegation model, architecture, server modes, token exchange, and encryption.',
  },
  {
    title: 'Get started',
    href: '/docs/get-started',
    description:
      'Run the full stack locally and walk a delegation from consent to token exchange.',
  },
  {
    title: 'API reference',
    href: '/api/enduser',
    description:
      'The end-user and admin OpenAPI contracts, rendered from the source specifications.',
  },
];

export default function HomepageFeatures() {
  return (
    <section className={styles.overviewSection}>
      <div className="container">
        <div className={styles.sectionHeader}>
          <span className={styles.sectionEyebrow}>What this is</span>
          <Heading as="h2" className={styles.sectionTitle}>
            Identity infrastructure for agents that act on a user&apos;s behalf
          </Heading>
          <p className={styles.sectionDescription}>
            These docs focus on the concerns an IAM team or platform operator
            cares about: the delegation model, security posture, deployment,
            configuration, and the API contracts other systems integrate against.
          </p>
        </div>

        <div className={styles.capabilityGrid}>
          {capabilityCards.map((card) => (
            <article key={card.title} className={styles.card}>
              <Heading as="h3" className={styles.cardTitle}>
                {card.title}
              </Heading>
              <p className={styles.cardDescription}>{card.description}</p>
            </article>
          ))}
        </div>

        <div className={styles.pathSection}>
          <div className={styles.sectionHeader}>
            <span className={styles.sectionEyebrow}>Start in the right place</span>
            <Heading as="h2" className={styles.sectionTitle}>
              Read by task, not by internals
            </Heading>
          </div>
          <div className={styles.pathGrid}>
            {pathCards.map((card) => (
              <Link key={card.href} className={styles.pathCard} to={card.href}>
                <span className={styles.pathTitle}>{card.title}</span>
                <p className={styles.pathDescription}>{card.description}</p>
                <span className={styles.pathAction}>Open section</span>
              </Link>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

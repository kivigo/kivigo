import type { ReactNode } from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HomepageFeatures from '@site/src/components/HomepageFeatures';
import Heading from '@theme/Heading';

import styles from './index.module.css';
function HomepageHeader() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className={`container ${styles.heroContent}`}>
        <img
          src="/kivigo/img/logo-kivigo.png"
          alt="KiviGo Logo"
          style={{ maxHeight: 140 }}
          className={styles.heroLogo}
        />
        <div className={styles.heroText}>
          <Heading as="h1" className="hero__title">
            {siteConfig.title}
          </Heading>
          <p className="hero__subtitle">{siteConfig.tagline}</p>
          <div className={styles.buttons}>
            <Link
              className="button button--secondary button--lg"
              to="/docs/getting-started">
              Get Started - 5min ⏱️
            </Link>
            <Link
              className="button button--outline button--lg"
              to="/docs/backends/overview"
              style={{ marginLeft: '1rem' }}>
              View Backends
            </Link>
          </div>
        </div>
      </div>
    </header>
  );
}

export default function Home(): ReactNode {
  const { siteConfig } = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title} - Lightweight key-value store library for Go`}
      description="KiviGo is a lightweight, modular key-value store library for Go that provides a unified interface for different backends like Redis, BadgerDB, etcd, Consul, and more.">
      <HomepageHeader />
      <main>
        <HomepageFeatures />
      </main>
    </Layout>
  );
}

import type { ReactNode } from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  Img: string; // chemin vers le PNG
  description: ReactNode;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Multiple Backends',
    Img: require('@site/static/img/kiwi-backends.png').default,
    description: (
      <>
        Support for Redis, BadgerDB, etcd, Consul, MongoDB, MySQL, PostgreSQL,
        and more. Switch between backends with minimal code changes and unified API.
      </>
    ),
  },
  {
    title: 'Built for Go',
    Img: require('@site/static/img/kiwi-go.png').default,
    description: (
      <>
        Native Go implementation with type safety, proper error handling,
        and context support. Designed following Go best practices and idioms.
      </>
    ),
  },
  {
    title: 'Enterprise Ready',
    Img: require('@site/static/img/kiwi-health.png').default,
    description: (
      <>
        Health checks, batch operations, custom encoders, mock testing support,
        and comprehensive documentation. Production-ready with monitoring capabilities.
      </>
    ),
  },
];

function Feature({ title, Img, description }: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <img src={Img} alt={title} className={styles.featureImg} />
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): ReactNode {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}

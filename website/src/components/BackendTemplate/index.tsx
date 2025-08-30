import React from 'react';
import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';
import styles from './styles.module.css';

interface Feature {
  name: string;
  supported: boolean;
  description?: string;
}

interface BackendTemplateProps {
  name: string;
  description: string;
  category: string;
  packageName: string;
  importPath: string;
  features: Feature[];
  dependencies?: string[];
  installationNotes?: string;
  configurationExample: string;
  usageExample: string;
  healthCheckExample?: string;
  batchExample?: string;
  notes?: string[];
  links?: Array<{
    text: string;
    url: string;
  }>;
}

function getFeatureGridClass(count: number) {
  if (count <= 4) return `${styles['features-compact-grid']} ${styles['grid-1']}`;
  if (count <= 8) return `${styles['features-compact-grid']} ${styles['grid-2']}`;
  return `${styles['features-compact-grid']} ${styles['grid-3']}`;
}

export default function BackendTemplate({
  name,
  description,
  category,
  packageName,
  importPath,
  features,
  dependencies = [],
  installationNotes,
  configurationExample,
  usageExample,
  healthCheckExample,
  batchExample,
  notes = [],
  links = []
}: BackendTemplateProps) {
  return (
    <div className={styles['backend-template']}>
      <div className={styles['backend-header']}>

        {/* container dédié pour les badges */}
        <div className={styles.badges}>
          <span className={`${styles.badge} ${styles['badge-light']}`}>{category}</span>
          <span className={`${styles.badge} ${styles['badge-green']}`}>{packageName}</span>
        </div>
      </div>

      <p className="backend-description">{description}</p>

      <h2>📦 Installation</h2>
      <CodeBlock language="bash">
        {`go get ${importPath}`}
      </CodeBlock>

      {installationNotes && (
        <div className="admonition admonition-info">
          <div className="admonition-content">
            <p>{installationNotes}</p>
          </div>
        </div>
      )}

      {dependencies.length > 0 && (
        <>
          <h3>Dependencies</h3>
          <ul>
            {dependencies.map((dep, index) => (
              <li key={index}>{dep}</li>
            ))}
          </ul>
        </>
      )}

      <h2>✨ Features</h2>
      <div className={getFeatureGridClass(features.length)}>
        {features.map((feature, index) => (
          <div
            key={index}
            className={`${styles['feature-compact-item']} ${feature.supported ? styles.supported : styles['not-supported']}`}
            title={feature.description || feature.name}
          >
            <span className={styles['feature-compact-status']}>
              {feature.supported ? '✅' : '❌'}
            </span>
            <span className={styles['feature-compact-name']}>
              {feature.name}
            </span>
          </div>
        ))}
      </div>

      <h2>🚀 Usage</h2>
      <Tabs>
        <TabItem value="configuration" label="Configuration" default>
          <CodeBlock language="go">
            {configurationExample}
          </CodeBlock>
        </TabItem>

        <TabItem value="basic-usage" label="Basic Usage">
          <CodeBlock language="go">
            {usageExample}
          </CodeBlock>
        </TabItem>

        {healthCheckExample && (
          <TabItem value="health-check" label="Health Check">
            <CodeBlock language="go">
              {healthCheckExample}
            </CodeBlock>
          </TabItem>
        )}

        {batchExample && (
          <TabItem value="batch-operations" label="Batch Operations">
            <CodeBlock language="go">
              {batchExample}
            </CodeBlock>
          </TabItem>
        )}
      </Tabs>

      {notes.length > 0 && (
        <>
          <h2>📝 Notes</h2>
          <ul>
            {notes.map((note, index) => (
              <li key={index}>{note}</li>
            ))}
          </ul>
        </>
      )}

      {links.length > 0 && (
        <>
          <h2>🔗 Additional Resources</h2>
          <ul>
            {links.map((link, index) => (
              <li key={index}>
                <a href={link.url} target="_blank" rel="noopener noreferrer">
                  {link.text}
                </a>
              </li>
            ))}
          </ul>
        </>
      )}
    </div>
  );
}

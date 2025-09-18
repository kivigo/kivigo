import React, { useEffect, useState } from 'react';
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
  // State for latest version
  const [latestVersion, setLatestVersion] = useState<string>('');
  const [installCmd, setInstallCmd] = useState<string>(`go get ${importPath}`);

  // Nettoyage du packageName pour ne garder que le nom du backend (ex: "local" à partir de "backend/local")
  function extractBackendName(pkg: string) {
    const parts = pkg.split('/');
    return parts[parts.length - 1];
  }

  // Fetch latest tag from GitHub on mount
  useEffect(() => {
    const backend = extractBackendName(packageName).toLowerCase();
    fetch('https://api.github.com/repos/kivigo/backends/tags?per_page=100')
      .then(res => res.json())
      .then((tags: { name: string }[]) => {
        // Filter tags for this backend
        const filtered = tags
          .map(tag => tag.name)
          .filter(tag => tag.startsWith(`${backend}/`))
          .map(tag => tag.split('/')[1]);
        if (filtered.length > 0) {
          // On prend la première version (GitHub renvoie les tags du plus récent au plus ancien)
          setLatestVersion(filtered[0]);
          setInstallCmd(`go get ${importPath}@${filtered[0]}`);
        }
      })
      .catch(() => {
        setLatestVersion('');
        setInstallCmd(`go get ${importPath}`);
      });
  }, [importPath, packageName]);

  // Lien vers la page des tags GitHub pour ce backend
  const backendName = extractBackendName(packageName).toLowerCase();
  const tagsUrl = `https://github.com/kivigo/backends/tags?q=${backendName}%2F`;

  return (
    <div className={styles['backend-template']}>
      <div className={styles['backend-header']}>
        <div className={styles.badges}>
          <span className={`${styles.badge} ${styles['badge-light']}`}>{category}</span>
          <span className={`${styles.badge} ${styles['badge-green']}`}>{packageName}</span>
        </div>
      </div>

      <p className="backend-description">{description}</p>

      <h2>📦 Installation</h2>
      <div className={styles.versionRow}>
        <span className={styles.versionLabel}>Latest version : {latestVersion}</span>
        <a
          href={tagsUrl}
          target="_blank"
          rel="noopener noreferrer"
          className={styles.versionTagsLink}
          style={{ marginLeft: 10, fontSize: 14 }}
        >
          View all versions
        </a>
      </div>
      <CodeBlock language="bash">
        {installCmd}
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
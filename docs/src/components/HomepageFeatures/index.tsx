import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import Link from '@docusaurus/Link';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  icon: string;
  description: ReactNode;
  link?: string;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Multi-Language Support',
    icon: '🌐',
    description: (
      <>
        Generate code for Go, TypeScript, Python, Java, and more. 
        Extend with custom plugins for any language or framework.
      </>
    ),
    link: '/docs/cli/generate'
  },
  {
    title: 'Registry System',
    icon: '📦',
    description: (
      <>
        Manage protobuf schemas in a single source of truth.
        Versioning, access control, and collaboration made easy.
      </>
    ),
    link: '/docs/registry/overview'
  },
  {
    title: 'Open Source',
    icon: '💝',
    description: (
      <>
        MIT licensed, community-driven, and extensible. Contribute to the project
        and help shape the future of protobuf tooling.
      </>
    ),
    link: 'https://github.com/sirrobot01/protodex'
  },
];

function Feature({title, icon, description, link}: FeatureItem) {
  const content = (
      <div className={styles.featureCard}>
          <div className={styles.featureIcon}>
          <span className={styles.icon} role="img" aria-label={title}>
            {icon}
          </span>
          </div>
          <div className={styles.featureContent}>
              <Heading as="h3" className={styles.featureTitle}>{title}</Heading>
              <p className={styles.featureDescription}>{description}</p>
          </div>
      </div>
  );

  if (link) {
    return (
      <Link to={link} className={styles.featureLink}>
        {content}
      </Link>
    );
  }

  return content;
}

function CodeExample() {
  return (
    <section className={`${styles.codeExample}`}>
      <div className="container">
        <div className="row">
          <div className="col col--6">
            <div className={styles.codeBlock}>
              <h3>Simple Configuration</h3>
              <pre className={styles.code}>
                <code>{`package:
  name: user-service
  description: User management service schemas

files:
  exclude: []
  base_dir: .

gen:
  languages:
    - name: go
      output_dir: ./gen/go
      module_path: github.com/myuser/user-service

deps:
  - name: google/protobuf
    source: google`}</code>
              </pre>
            </div>
          </div>
          <div className="col col--6">
            <div className={styles.commandExample}>
              <h3>Quick Commands</h3>
              <div className={styles.terminalBlock}>
                <div className={styles.terminalHeader}>
                  <span className={styles.terminalButton}></span>
                  <span className={styles.terminalButton}></span>
                  <span className={styles.terminalButton}></span>
                </div>
                <pre className={styles.terminal}>
                  <code>{`$ protodex init
$ protodex validate
All schemas valid

$ protodex generate go
Generated Go code
Generated TypeScript code

$ protodex push
Published to registry`}</code>
                </pre>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

export default function HomepageFeatures(): ReactNode {
  return (
    <>
      <section className={styles.features}>
        <div className="container">
          <div className="row">
            {FeatureList.map((props, idx) => (
              <div className="col col--4" key={idx}>
                  <Feature key={idx} {...props} />
              </div>
            ))}
          </div>
        </div>
      </section>
      <CodeExample />
    </>
  );
}

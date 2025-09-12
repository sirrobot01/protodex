import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HomepageFeatures from '@site/src/components/HomepageFeatures';
import Heading from '@theme/Heading';
import { useEffect, useRef } from 'react'; // Add this import
import { annotate } from 'rough-notation'; // Add this import

import styles from './index.module.css';

function HomepageHeader() {
    const {siteConfig} = useDocusaurusContext();
    const protodexRef = useRef(null); // Add this ref

    // Add this useEffect
    useEffect(() => {
        if (protodexRef.current) {
            const annotation = annotate(protodexRef.current, {
                type: 'underline',
                color: '#FF5733',
                strokeWidth: 3,
                padding: 2,
                iterations: 2,
                multiline: false,
                animationDuration: 800
            });

            // Show annotation after a small delay for better UX
            setTimeout(() => {
                annotation.show();
            }, 500);
        }
    }, []);

    return (
        <header className={clsx('hero', styles.heroBanner)}>
            <div className="container">
                <div className={styles.heroContent}>
                    <div className={styles.heroText}>
                        <Heading as="h1" className={styles.heroTitle}>
                            Protodex, an open source <span ref={protodexRef}>Protobuf</span> Toolchain
                        </Heading>
                        <p className={styles.heroSubtitle}>
                            The fastest way to manage Protocol Buffer schemas, generate code for multiple languages and manage packages across teams.
                        </p>
                        <div className={styles.heroButtons}>
                            <Link
                                className={clsx('button', styles.primaryButton)}
                                to="/docs/intro">
                                Get Started
                            </Link>
                            <Link
                                className={clsx('button', styles.secondaryButton)}
                                to="/docs/quick-start">
                                Quick Start →
                            </Link>
                        </div>
                    </div>
                    {/* Rest of your component stays the same */}
                    <div className={styles.heroVisual}>
                        <div className={styles.codePreview}>
                            <div className={styles.codeHeader}>
                                <div className={styles.codeTabs}>
                                    <span className={styles.codeTab}>protodex.yaml</span>
                                </div>
                            </div>
                            <div className={styles.codeContent}>
                <pre>{`package:
  name: user-service
  description: User management service

files:
  exclude: []
  base_dir: .

gen:
  languages:
    - name: go
      output_dir: ./gen/go
      module_path: github.com/user/user

deps:
  - name: google/protobuf
    source: google`}</pre>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </header>
    );
}

// Rest of your component stays exactly the same
function CallToAction() {
    return (
        <section className={styles.cta}>
            <div className="container">
                <div className={styles.ctaContent}>
                    <Heading as="h2" className={styles.ctaTitle}>
                        Ready to modernize your protobuf workflow?
                    </Heading>
                    <div className={styles.ctaButtons}>
                        <Link
                            className={clsx('button', styles.primaryButton)}
                            to="/docs/getting-started/installation">
                            Install Protodex
                        </Link>
                        <Link
                            className={clsx('button', styles.githubButton)}
                            to="https://github.com/sirrobot01/protodex">
                            ⭐ Star on GitHub
                        </Link>
                    </div>
                </div>
            </div>
        </section>
    );
}

export default function Home(): ReactNode {
    const {siteConfig} = useDocusaurusContext();
    return (
        <Layout
            title="Protodex, a modern Protobuf Toolchain"
            description="The fastest way to manage Protocol Buffer schemas, generate code for multiple languages, and share packages across teams.">
            <HomepageHeader />
            <main>
                <HomepageFeatures />
                <CallToAction />
            </main>
        </Layout>
    );
}
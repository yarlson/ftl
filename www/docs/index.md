---
layout: home
title: FTL - Simple Docker Deployment Tool for Developers
description: Deploy web applications easily without DevOps expertise. Free SSL, database management, and zero-downtime updates. Simple alternative to Kamal, Sidekick and complex deployment tools.
head:
  - - meta
    - name: keywords
      content: docker deployment, web app deployment, kamal alternative, sidekick alternative, zero-downtime deployment, SSL automation, database provisioning, simple deployment tool, deploy without devops, easy deployment
  - - meta
    - name: og:title
      content: FTL - Simple Docker Deployment Tool for Developers
  - - meta
    - name: og:description
      content: Deploy web applications easily without DevOps expertise. Free SSL, database management, and zero-downtime updates. Simple alternative to Kamal, Sidekick and complex deployment tools.
  - - meta
    - name: og:type
      content: website
  - - meta
    - name: twitter:title
      content: FTL - Simple Docker Deployment Tool for Developers
  - - meta
    - name: twitter:description
      content: Deploy web applications easily without DevOps expertise. Free SSL, database management, and zero-downtime updates. Simple alternative to Kamal, Sidekick and complex deployment tools.
---

<style>
.home-container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 40px;
  align-items: start;
  margin-top: 2rem;
}

.feature-list {
  list-style: none;
  padding: 0;
}

.feature-list li {
  margin-bottom: 1rem;
  font-size: 1.1em;
}

.quick-links {
  margin-top: 2rem;
}

.quick-links-list {
  list-style: none;
  padding: 0;
}

.quick-links-list li {
  margin-bottom: 1rem;
}

.quick-links-list a {
  font-size: 1.1em;
  font-weight: 500;
}

.quick-links-list span {
  display: block;
  margin-top: 0.25rem;
  color: var(--vp-c-text-2);
}

.home-content {
    margin: 16px;
}
</style>

<div class="home-container">
<div class="home-content">

# FTL: Faster Than Light Deployment

Deploy your web applications to production servers without complexity. FTL handles everything from server setup to SSL certificates.

<div class="feature-list">

🚀 Deploy with a single command

🔒 Automatic server security setup

🌐 Free SSL certificates included

📦 Database setup and management

♻️ Updates without downtime

🛠️ No DevOps expertise needed

</div>

<div class="quick-links">

## Get Started in Minutes

<ul class="quick-links-list">
  <li>
    <a href="/getting-started/installation">Install FTL</a>
    <span>One command to install on your computer</span>
  </li>
  <li>
    <a href="/getting-started/first-deployment">Deploy Your First App</a>
    <span>Step-by-step guide to your first deployment</span>
  </li>
  <li>
    <a href="/guides/github-actions">GitHub Actions Integration</a>
    <span>Automate deployments with GitHub Actions</span>
  </li>
  <li>
    <a href="/guides/concepts">Core Concepts</a>
    <span>Learn the basics of deployment</span>
  </li>
  <li>
    <a href="/examples/">Example Projects</a>
    <span>Ready-to-use deployment examples</span>
  </li>
</ul>

</div>

</div>
<div class="code">

```yaml
# Simple configuration - just fill in your details
project:
  name: my-website # Your project name
  domain: mysite.com # Your domain name
  email: me@mysite.com # Your email for SSL

# Your server details from your hosting provider
server:
  host: 64.23.132.12 # Your server IP
  user: deploy
  ssh_key: ~/.ssh/id_rsa

# Your application
services:
  - name: website
    port: 3000 # Your app's port

# Need a database? Just add it here
dependencies:
  - postgres:16
```

</div>
</div>

## What is FTL?

FTL helps developers deploy web applications to production servers. It automates all the complex parts of deployment that usually require DevOps expertise:

- Sets up your server with all required security
- Installs and configures your database
- Gets free SSL certificates for your domain
- Keeps your site running during updates

### How Simple Is It?

1. Install FTL on your computer
2. Rent a basic server from any provider
3. Create a simple config file
4. Run `ftl deploy`

That's it. FTL handles everything else - server setup, security, SSL, databases, and more.

### Perfect For

- Developers deploying their first production application
- Small to medium web applications
- Teams without dedicated DevOps engineers
- Anyone who wants to focus on coding, not server management

### When to Consider Alternatives

- Large enterprise applications
- Microservice architectures with many components

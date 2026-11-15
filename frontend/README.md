# TEX TALK Frontend

Blank frontend project structure for the TEX TALK website.

## Stack

- Angular 19 standalone components
- TypeScript 5.6
- SCSS with Tailwind CSS v4 through PostCSS
- Ant Design (ng-zorro-antd) dependency and global stylesheet ready
- npm package manager

## Scripts

```bash
npm install
npm start
npm run build
npm run lint
npm run format
npm run format:lint
```

## Structure

```text
public/
  assets/
  env/
src/
  app/
    components/    Reusable UI components
    core/          Constants, enums, guards, interceptors, interfaces, models, pipes, utils
    directives/    Custom Angular directives
    pages/         Route-level pages
    services/      Business logic and API services
    shared/        Shared modules and utilities
  index.html
  main.ts
  styles.scss
  theme.less
```

No landing page, registration form, mock service, or feature code is included yet.
Pipeline files are intentionally not included yet.

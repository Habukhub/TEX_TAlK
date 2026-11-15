import { Routes } from '@angular/router';

import { AdminPage } from './pages/admin/admin.page';
import { AdminLoginPage } from './pages/admin-login/admin-login.page';
import { HomePage } from './pages/home/home.page';
import { adminAuthGuard } from './guards/admin-auth.guard';

export const routes: Routes = [
  {
    path: '',
    component: HomePage,
    pathMatch: 'full',
  },
  {
    path: 'admin/login',
    component: AdminLoginPage,
  },
  {
    path: 'admin',
    component: AdminPage,
    canActivate: [adminAuthGuard],
  },
];

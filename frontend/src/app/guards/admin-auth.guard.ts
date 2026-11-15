import { inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { CanActivateFn, Router } from '@angular/router';
import { catchError, map, of } from 'rxjs';
import { environment } from '../../environments/environment';

interface AdminJwtPayload {
  exp?: number;
  email?: string;
  user_id?: string;
}

export const adminAuthGuard: CanActivateFn = () => {
  const router = inject(Router);
  const http = inject(HttpClient);
  const token = localStorage.getItem('jwt_token');

  if (!isTokenUsable(token)) {
    return redirectToLogin(router);
  }

  return http.get(`${environment.apiUrl}/admin/submissions/pending`).pipe(
    map(() => true),
    catchError(() => of(redirectToLogin(router))),
  );
};

function redirectToLogin(router: Router) {
  localStorage.removeItem('jwt_token');
  return router.createUrlTree(['/admin/login']);
}

function isTokenUsable(token: string | null): boolean {
  if (!token) {
    return false;
  }

  const payload = readJwtPayload(token);
  return Boolean(
    typeof payload?.exp === 'number' &&
    payload.exp * 1000 > Date.now() &&
    payload.email &&
    payload.user_id,
  );
}

function readJwtPayload(token: string): AdminJwtPayload | null {
  const tokenParts = token.split('.');
  if (tokenParts.length !== 3) {
    return null;
  }

  const [, payload] = tokenParts;
  if (!payload) {
    return null;
  }

  try {
    const normalizedPayload = payload.replace(/-/g, '+').replace(/_/g, '/');
    const paddedPayload = normalizedPayload.padEnd(
      normalizedPayload.length + ((4 - (normalizedPayload.length % 4)) % 4),
      '=',
    );
    return JSON.parse(atob(paddedPayload)) as AdminJwtPayload;
  } catch {
    return null;
  }
}

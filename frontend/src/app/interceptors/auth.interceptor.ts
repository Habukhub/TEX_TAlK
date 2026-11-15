import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const router = inject(Router);

  // Check if it's an API request
  if (req.url.includes('/api/')) {
    const token = localStorage.getItem('jwt_token');

    if (token) {
      req = req.clone({
        setHeaders: {
          Authorization: `Bearer ${token}`,
        },
      });
    }
  }

  return next(req).pipe(
    catchError((error) => {
      if (req.url.includes('/admin/') && (error.status === 401 || error.status === 403)) {
        localStorage.removeItem('jwt_token');
        router.navigate(['/admin/login'], { replaceUrl: true });
      }

      return throwError(() => error);
    }),
  );
};

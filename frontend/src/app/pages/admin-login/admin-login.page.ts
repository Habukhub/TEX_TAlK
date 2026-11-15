import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, ActivatedRoute } from '@angular/router';
import { environment } from '../../../environments/environment';

@Component({
  selector: 'app-admin-login-page',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './admin-login.page.html',
  styleUrl: './admin-login.page.scss',
})
export class AdminLoginPage implements OnInit {
  private router = inject(Router);
  private route = inject(ActivatedRoute);

  alertMessage = '';
  alertType: 'success' | 'error' | '' = '';

  ngOnInit(): void {
    // Check if we just returned from Google OAuth
    this.route.queryParams.subscribe(params => {
      const token = params['token'];
      const error = params['error'];

      if (token) {
        localStorage.setItem('jwt_token', token);
        this.alertType = 'success';
        this.alertMessage = 'Login successful. Welcome!';
        // Navigate to dashboard after showing message
        setTimeout(() => {
          this.router.navigate(['/admin'], { replaceUrl: true });
        }, 1500);
      } else if (error === 'unauthorized_email') {
        this.alertType = 'error';
        this.alertMessage = 'Access denied. This email is not registered as an Admin.';
        // Clean up the URL
        this.router.navigate(['/admin/login'], { replaceUrl: true });
      }
    });
  }

  protected loginWithGoogle(): void {
    // Redirect to backend Google OAuth endpoint
    window.location.href = `${environment.apiUrl}/auth/google`;
  }
}

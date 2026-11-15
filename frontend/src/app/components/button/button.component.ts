import { Component, Input, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-button',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './button.component.html',
  styleUrl: './button.component.scss'
})
export class ButtonComponent {
  @Input() variant: 'home-primary' | 'home-secondary' | 'admin-primary' | 'admin-secondary' | 'admin-approve' | 'admin-reject' | 'admin-edit' | 'admin-edit-blue' | 'admin-icon' = 'home-primary';
  @Input() type: 'button' | 'submit' | 'reset' = 'button';
  @Input() fullWidth: boolean = false;
  @Input() disabled: boolean = false;

  @Output() clicked = new EventEmitter<Event>();

  get buttonClass(): string {
    let classes = '';
    
    // Home variants
    if (this.variant === 'home-primary') classes = 'home-button home-button--primary';
    if (this.variant === 'home-secondary') classes = 'home-button home-button--secondary';
    
    // Admin variants
    if (this.variant === 'admin-primary') classes = 'primary-button';
    if (this.variant === 'admin-secondary') classes = 'secondary-button';
    if (this.variant === 'admin-approve') classes = 'small-button approve';
    if (this.variant === 'admin-reject') classes = 'small-button reject';
    if (this.variant === 'admin-edit') classes = 'small-button edit';
    if (this.variant === 'admin-edit-blue') classes = 'small-button edit-blue';
    if (this.variant === 'admin-icon') classes = 'icon-button';

    if (this.fullWidth) {
      classes += ' home-button--full';
    }

    return classes;
  }
}

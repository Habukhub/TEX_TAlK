import { Component, Input, Output, EventEmitter } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-modal',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './modal.component.html',
  styleUrl: './modal.component.scss'
})
export class ModalComponent {
  @Input() isOpen = false;
  @Input() kicker = '';
  @Input() title = '';
  @Output() dismissed = new EventEmitter<void>();

  onClose() {
    this.dismissed.emit();
  }
}

import { Component, ElementRef, EventEmitter, Input, OnChanges, Output, SimpleChanges, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ButtonComponent } from '../button/button.component';
import { ModalComponent } from '../modal/modal.component';
import { PublicSpeaker } from '../../services/application.service';

export interface PresentationUploadData {
  speakerId: string;
  speakerName: string;
  topic: string;
  file: File;
}

@Component({
  selector: 'app-upload-presentation',
  standalone: true,
  imports: [CommonModule, FormsModule, ButtonComponent, ModalComponent],
  templateUrl: './upload-presentation.component.html',
  styleUrl: './upload-presentation.component.scss'
})
export class UploadPresentationComponent implements OnChanges {
  @Input() isOpen = false;
  @Input() speakers: PublicSpeaker[] = [];
  @Output() dismissed = new EventEmitter<void>();
  @Output() upload = new EventEmitter<PresentationUploadData>();
  @ViewChild('presentationFileInput') presentationFileInput?: ElementRef<HTMLInputElement>;

  selectedSpeakerId = '';
  selectedSpeakerName = '';
  selectedTopic = '';
  selectedFile: File | null = null;
  isSpeakerDropdownOpen = false;

  ngOnChanges(changes: SimpleChanges) {
    if (changes['isOpen']?.currentValue === true) {
      this.resetForm();
    }
  }

  onSpeakerChange() {
    const speaker = this.speakers.find(s => s.full_name === this.selectedSpeakerName);
    if (speaker) {
      this.selectedSpeakerId = speaker.id;
      this.selectedTopic = speaker.topic_title || '';
    } else {
      this.selectedSpeakerId = '';
      this.selectedTopic = '';
    }
  }

  openSpeakerDropdown() {
    this.isSpeakerDropdownOpen = true;
  }

  closeSpeakerDropdown() {
    window.setTimeout(() => {
      this.isSpeakerDropdownOpen = false;
    }, 120);
  }

  selectSpeaker(speaker: PublicSpeaker) {
    this.selectedSpeakerName = speaker.full_name;
    this.selectedSpeakerId = speaker.id;
    this.selectedTopic = speaker.topic_title || '';
    this.isSpeakerDropdownOpen = false;
  }

  get filteredSpeakers() {
    const query = this.selectedSpeakerName.trim().toLowerCase();
    if (!query) {
      return this.speakers;
    }

    return this.speakers.filter(speaker =>
      speaker.full_name?.toLowerCase().includes(query),
    );
  }

  readonly allowedMimeTypes = [
    'application/pdf',
    'application/vnd.ms-powerpoint',
    'application/vnd.openxmlformats-officedocument.presentationml.presentation',
  ];
  readonly allowedExtensions = ['.pdf', '.ppt', '.pptx'];

  onFileChange(event: Event) {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0] ?? null;
    if (file) {
      const ext = file.name.slice(file.name.lastIndexOf('.')).toLowerCase();
      if (!this.allowedExtensions.includes(ext) && !this.allowedMimeTypes.includes(file.type)) {
        alert('รองรับเฉพาะไฟล์ PDF, PPT, PPTX เท่านั้น');
        input.value = '';
        this.selectedFile = null;
        return;
      }
      const maxSize = 50 * 1024 * 1024;
      if (file.size > maxSize) {
        alert('ไฟล์มีขนาดใหญ่เกินไป! กรุณาอัปโหลดไฟล์ที่มีขนาดไม่เกิน 50MB ครับ');
        input.value = '';
        this.selectedFile = null;
        return;
      }
    }
    this.selectedFile = file ?? null;
  }

  onClose() {
    this.dismissed.emit();
    this.resetForm();
  }

  onSubmit() {
    if (!this.selectedSpeakerName.trim() || !this.selectedFile) {
      alert('Please enter a speaker and select a file.');
      return;
    }

    this.upload.emit({
      speakerId: this.selectedSpeakerId,
      speakerName: this.selectedSpeakerName.trim(),
      topic: this.selectedTopic,
      file: this.selectedFile
    });
    
    this.onClose();
  }

  private resetForm() {
    this.selectedSpeakerId = '';
    this.selectedSpeakerName = '';
    this.selectedTopic = '';
    this.selectedFile = null;
    this.isSpeakerDropdownOpen = false;
    if (this.presentationFileInput) {
      this.presentationFileInput.nativeElement.value = '';
    }
  }
}

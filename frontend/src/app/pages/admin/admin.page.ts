import { CommonModule } from '@angular/common';
import { Component, computed, inject, signal, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { ButtonComponent } from '../../components/button/button.component';
import { ApplicationService, SpeakerApplication, TimelineEntry } from '../../services/application.service';

interface EditInfoForm {
  applicantName: string;
  topic: string;
}

type AdminTab = 'pending' | 'history' | 'timeline' | 'settings';

interface ScheduleForm {
  applicationId: string | null;
  topic: string;
  startTime: string;
  endTime: string;
}

@Component({
  selector: 'app-admin-page',
  imports: [CommonModule, FormsModule, ButtonComponent],
  templateUrl: './admin.page.html',
  styleUrl: './admin.page.scss',
})
export class AdminPage implements OnInit {
  protected readonly activeTab = signal<AdminTab>('pending');
  protected readonly isSidebarOpen = signal(false);
  protected readonly isSidebarCollapsed = signal(false);
  protected readonly isScheduleModalOpen = signal(false);
  protected readonly isFeedbackModalOpen = signal(false);
  protected readonly isDeleteConfirmModalOpen = signal(false);
  protected readonly isRejectConfirmModalOpen = signal(false);
  protected readonly isToastOpen = signal(false);
  protected readonly editingEntryId = signal<string | null>(null);
  protected readonly pendingDeleteEntryId = signal<string | null>(null);
  protected readonly pendingRejectApplicationId = signal<string | null>(null);
  protected readonly isDeleteAppConfirmModalOpen = signal(false);
  protected readonly pendingDeleteAppId = signal<string | null>(null);
  protected readonly formError = signal('');
  protected readonly feedbackTitle = signal('');
  protected readonly feedbackMessage = signal('');
  protected readonly feedbackTone = signal<'success' | 'error'>('success');
  protected readonly toastTitle = signal('');
  protected readonly toastMessage = signal('');
  protected scheduleForm: ScheduleForm = this.createEmptyScheduleForm();
  protected selectedForZip = new Set<string>();

  // Edit Info Modal
  protected readonly isEditInfoModalOpen = signal(false);
  protected readonly editingInfoApplicationId = signal<string | null>(null);
  protected editInfoForm: EditInfoForm = { applicantName: '', topic: '' };

  protected readonly applicationService = inject(ApplicationService);
  private readonly router = inject(Router);
  private toastTimeoutId: ReturnType<typeof setTimeout> | null = null;

  protected readonly applications = this.applicationService.applications;
  protected readonly timelineEntries = this.applicationService.timelineEntries;
  protected readonly isPresentationUploadEnabled = this.applicationService.isPresentationUploadEnabled;

  ngOnInit(): void {
    this.applicationService.loadAdminData();
  }

  protected readonly pendingApplications = computed(() =>
    this.applications().filter((application) => application.status === 'pending'),
  );

  protected readonly approvedApplications = computed(() =>
    this.applications().filter((application) => application.status === 'approved'),
  );

  protected readonly scheduledApplicationIds = computed(
    () => new Set(this.timelineEntries().map((entry) => entry.applicationId)),
  );

  protected readonly availableApprovedApplications = computed(() =>
    this.approvedApplications().filter((application) => {
      const editingEntry = this.timelineEntries().find(
        (entry) => entry.id === this.editingEntryId(),
      );

      return (
        application.id === editingEntry?.applicationId ||
        !this.scheduledApplicationIds().has(application.id)
      );
    }),
  );

  protected readonly stats = computed(() => ({
    total: this.applications().length,
    pending: this.pendingApplications().length,
    approved: this.approvedApplications().length,
    scheduled: this.timelineEntries().length,
  }));

  protected readonly scheduleModalTitle = computed(() =>
    this.editingEntryId() ? 'Edit Slot' : 'Add Slot',
  );

  protected readonly scheduleModalDescription = computed(() =>
    this.editingEntryId() ? 'Update speaker and time.' : 'Choose a speaker and time.',
  );

  protected setActiveTab(tab: AdminTab): void {
    this.activeTab.set(tab);
    if (this.isMobileViewport()) {
      this.isSidebarOpen.set(false);
    }
  }

  protected toggleSidebar(): void {
    if (this.isMobileViewport()) {
      this.isSidebarOpen.update((isOpen) => !isOpen);
      return;
    }

    this.isSidebarCollapsed.update((isCollapsed) => !isCollapsed);
  }

  protected closeSidebar(): void {
    this.isSidebarOpen.set(false);
  }

  protected isSidebarExpanded(): boolean {
    return this.isMobileViewport() ? this.isSidebarOpen() : !this.isSidebarCollapsed();
  }

  protected approveApplication(applicationId: string): void {
    this.applicationService.approveApplication(applicationId).then(() => {
      this.openToast(
        'Application approved',
        'The submission has been approved successfully.',
      );
    }).catch(() => {
      this.openFeedbackModal(
        'error',
        'Approve Failed',
        'The submission could not be approved. Please try again.',
      );
    });
  }

  protected rejectApplication(applicationId: string): void {
    this.pendingRejectApplicationId.set(applicationId);
    this.isRejectConfirmModalOpen.set(true);
  }

  protected closeRejectConfirmModal(): void {
    this.isRejectConfirmModalOpen.set(false);
    this.pendingRejectApplicationId.set(null);
  }

  protected confirmRejectApplication(): void {
    const applicationId = this.pendingRejectApplicationId();
    if (!applicationId) {
      return;
    }

    const linkedEntry = this.timelineEntries().find(e => e.applicationId === applicationId);

    this.applicationService.rejectApplication(applicationId).then(() => {
      if (linkedEntry && linkedEntry.id) {
        this.applicationService.deleteTimelineEntry(linkedEntry.id).catch(err => console.error('Failed to clean up timeline', err));
      }
      this.closeRejectConfirmModal();
      this.openToast(
        'Application rejected',
        'The submission has been rejected successfully.',
      );
    }).catch(() => {
      this.closeRejectConfirmModal();
      this.openFeedbackModal(
        'error',
        'Reject Failed',
        'The submission could not be rejected. Please try again.',
      );
    });
  }

  protected deleteApplicationHistory(applicationId: string): void {
    this.pendingDeleteAppId.set(applicationId);
    this.isDeleteAppConfirmModalOpen.set(true);
  }

  protected closeDeleteAppConfirmModal(): void {
    this.isDeleteAppConfirmModalOpen.set(false);
    this.pendingDeleteAppId.set(null);
  }

  protected confirmDeleteApplicationHistory(): void {
    const applicationId = this.pendingDeleteAppId();
    if (!applicationId) return;

    const linkedEntry = this.timelineEntries().find(e => e.applicationId === applicationId);

    this.applicationService.rejectApplication(applicationId).then(() => {
      if (linkedEntry && linkedEntry.id) {
        this.applicationService.deleteTimelineEntry(linkedEntry.id).catch(err => console.error('Failed to clean up timeline', err));
      }
      this.closeDeleteAppConfirmModal();
      this.openToast(
        'Application deleted',
        'The submission has been deleted successfully.',
      );
    }).catch(() => {
      this.closeDeleteAppConfirmModal();
      this.openFeedbackModal(
        'error',
        'Delete Failed',
        'The submission could not be deleted. Please try again.',
      );
    });
  }

  protected openEditInfoModal(application: SpeakerApplication): void {
    this.editingInfoApplicationId.set(application.id);
    this.editInfoForm = { applicantName: application.applicantName, topic: application.topic };
    this.isEditInfoModalOpen.set(true);
  }

  protected closeEditInfoModal(): void {
    this.isEditInfoModalOpen.set(false);
    this.editingInfoApplicationId.set(null);
    this.editInfoForm = { applicantName: '', topic: '' };
  }

  protected saveEditInfo(): void {
    const applicationId = this.editingInfoApplicationId();
    if (!applicationId) return;

    if (!this.editInfoForm.applicantName.trim() || !this.editInfoForm.topic.trim()) {
      return;
    }

    this.applicationService
      .updateSubmissionInfo(applicationId, this.editInfoForm.applicantName, this.editInfoForm.topic)
      .then(() => {
        this.closeEditInfoModal();
        this.openToast('Info updated', 'The applicant info has been updated successfully.');
      })
      .catch(() => {
        this.openFeedbackModal('error', 'Update Failed', 'The submission info could not be updated. Please try again.');
      });
  }

  protected downloadSlide(applicationId: string): void {
    this.applicationService.downloadSlide(applicationId);
  }

  protected downloadZip(): void {
    if (this.selectedForZip.size === 0) return;
    this.applicationService.downloadZip(Array.from(this.selectedForZip));
  }

  protected toggleZipSelection(applicationId: string): void {
    const nextSet = new Set(this.selectedForZip);
    if (nextSet.has(applicationId)) {
      nextSet.delete(applicationId);
    } else {
      nextSet.add(applicationId);
    }
    this.selectedForZip = nextSet;
  }

  protected toggleAllZips(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (input.checked) {
      const allIds = this.approvedApplications()
        .filter(app => app.slideFileName)
        .map(app => app.id);
      this.selectedForZip = new Set(allIds);
    } else {
      this.selectedForZip = new Set();
    }
  }

  protected areAllZipsSelected(): boolean {
    const withSlides = this.approvedApplications().filter(app => app.slideFileName);
    if (withSlides.length === 0) return false;
    return this.selectedForZip.size === withSlides.length;
  }

  protected logout(): void {
    localStorage.removeItem('jwt_token');
    this.router.navigate(['/admin/login']);
  }

  protected openScheduleModal(): void {
    this.editingEntryId.set(null);
    this.scheduleForm = this.createEmptyScheduleForm();
    this.formError.set('');
    this.isScheduleModalOpen.set(true);
  }

  protected closeScheduleModal(): void {
    this.isScheduleModalOpen.set(false);
    this.editingEntryId.set(null);
    this.formError.set('');
  }

  protected closeFeedbackModal(): void {
    this.isFeedbackModalOpen.set(false);
  }

  protected closeDeleteConfirmModal(): void {
    this.isDeleteConfirmModalOpen.set(false);
    this.pendingDeleteEntryId.set(null);
  }

  protected editScheduleEntry(entry: TimelineEntry): void {
    this.editingEntryId.set(entry.id);
    this.scheduleForm = {
      applicationId: entry.applicationId ?? null,
      topic: entry.topic,
      startTime: entry.startTime,
      endTime: entry.endTime,
    };
    this.formError.set('');
    this.isScheduleModalOpen.set(true);
  }

  protected updateSelectedSpeaker(): void {
    const selectedApplication = this.findApplication(this.scheduleForm.applicationId);
    this.scheduleForm.topic = selectedApplication?.topic ?? '';
  }

  protected saveScheduleEntry(): void {
    const selectedApplication = this.findApplication(this.scheduleForm.applicationId);

    if (!selectedApplication) {
      this.formError.set('Please select an approved speaker.');
      return;
    }

    if (!this.scheduleForm.startTime || !this.scheduleForm.endTime) {
      this.formError.set('Please complete start time and end time.');
      return;
    }

    if (this.scheduleForm.startTime >= this.scheduleForm.endTime) {
      this.formError.set('End time must be after start time.');
      return;
    }

    const newStart = this.timeToMinutes(this.scheduleForm.startTime);
    const newEnd = this.timeToMinutes(this.scheduleForm.endTime);
    const editingId = this.editingEntryId();

    const hasOverlap = this.timelineEntries().some((entry) => {
      // Skip checking against the entry currently being edited
      if (editingId && entry.id === editingId) {
        return false;
      }

      const existingStart = this.timeToMinutes(entry.startTime);
      const existingEnd = this.timeToMinutes(entry.endTime);

      // Two intervals overlap if (StartA < EndB) and (EndA > StartB)
      return newStart < existingEnd && newEnd > existingStart;
    });

    if (hasOverlap) {
      this.formError.set('This time overlaps with another slot.');
      return;
    }

    const isEdit = !!this.editingEntryId();
    const nextEntry: TimelineEntry = {
      id: this.editingEntryId() ?? Date.now().toString(),
      applicationId: selectedApplication.id,
      speaker: selectedApplication.applicantName,
      topic: selectedApplication.topic,
      startTime: this.scheduleForm.startTime,
      endTime: this.scheduleForm.endTime,
    };

    this.applicationService.saveTimelineEntry(nextEntry, isEdit).then(() => {
      this.closeScheduleModal();
      this.openToast(
        isEdit ? 'Timeline updated' : 'Timeline created',
        isEdit
          ? 'The timeline entry has been updated successfully.'
          : 'The timeline entry has been created successfully.',
      );
    }).catch(() => {
      this.formError.set(
        isEdit ? 'Failed to update the timeline entry.' : 'Failed to create the timeline entry.',
      );
      this.openFeedbackModal(
        'error',
        isEdit ? 'Update Failed' : 'Create Failed',
        isEdit
          ? 'The timeline entry could not be updated. Please try again.'
          : 'The timeline entry could not be created. Please try again.',
      );
    });
  }

  protected deleteTimelineEntry(entryId: string): void {
    this.pendingDeleteEntryId.set(entryId);
    this.isDeleteConfirmModalOpen.set(true);
  }

  protected confirmDeleteTimelineEntry(): void {
    const entryId = this.pendingDeleteEntryId();
    if (!entryId) {
      return;
    }

    this.applicationService.deleteTimelineEntry(entryId).then(() => {
      this.closeDeleteConfirmModal();
      this.openToast(
        'Timeline deleted',
        'The timeline entry has been deleted successfully.',
      );
    }).catch(() => {
      this.closeDeleteConfirmModal();
      this.openFeedbackModal(
        'error',
        'Delete Failed',
        'The timeline entry could not be deleted. Please try again.',
      );
    });
  }

  protected setPresentationUploadEnabled(event: Event): void {
    const input = event.target as HTMLInputElement;
    this.applicationService.setPresentationUploadEnabled(input.checked);
  }

  protected getDuration(entry: TimelineEntry): string {
    const start = this.timeToMinutes(entry.startTime);
    const end = this.timeToMinutes(entry.endTime);
    return `${end - start} min`;
  }

  protected trackById(_: number, item: SpeakerApplication | TimelineEntry): string {
    return item.id;
  }

  private findApplication(applicationId: string | null): SpeakerApplication | undefined {
    if (!applicationId) {
      return undefined;
    }

    return this.applications().find((application) => application.id === applicationId);
  }

  private timeToMinutes(time: string): number {
    const [hours, minutes] = time.split(':').map(Number);
    return hours * 60 + minutes;
  }

  private createEmptyScheduleForm(): ScheduleForm {
    return {
      applicationId: null,
      topic: '',
      startTime: '',
      endTime: '',
    };
  }

  private isMobileViewport(): boolean {
    return window.matchMedia('(max-width: 900px)').matches;
  }

  private openFeedbackModal(
    tone: 'success' | 'error',
    title: string,
    message: string,
  ): void {
    this.feedbackTone.set(tone);
    this.feedbackTitle.set(title);
    this.feedbackMessage.set(message);
    this.isFeedbackModalOpen.set(true);
  }

  protected closeToast(): void {
    this.isToastOpen.set(false);
    if (this.toastTimeoutId) {
      clearTimeout(this.toastTimeoutId);
      this.toastTimeoutId = null;
    }
  }

  private openToast(title: string, message: string): void {
    this.toastTitle.set(title);
    this.toastMessage.set(message);
    this.isToastOpen.set(true);

    if (this.toastTimeoutId) {
      clearTimeout(this.toastTimeoutId);
    }

    this.toastTimeoutId = setTimeout(() => {
      this.isToastOpen.set(false);
      this.toastTimeoutId = null;
    }, 2800);
  }
}

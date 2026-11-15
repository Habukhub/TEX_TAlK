import { Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map, catchError, of } from 'rxjs';
import { environment } from '../../environments/environment';

function devLog(...args: unknown[]): void {
  if (!environment.production) {
    console.error(...args);
  }
}

export type ApplicationStatus = 'pending' | 'approved' | 'rejected';

export interface SpeakerApplication {
  id: string;
  applicantName: string;
  email: string;
  topic: string;
  submissionDate: string;
  status: ApplicationStatus;
  approvedAt?: string;
  slideFileName?: string;
}

export interface TimelineEntry {
  id: string;
  applicationId?: string;
  speaker: string;
  topic: string;
  startTime: string;
  endTime: string;
}

export interface PublicSpeaker {
  id: string;
  full_name: string;
  topic_title: string;
  status: string;
  slide_file_name?: string | null;
}

interface BackendSchedule {
  ID: string;
  SubmissionID?: string | null;
  Submission?: {
    FullName?: string;
    TopicTitle?: string;
  } | null;
  ScheduleType?: string;
  StartTime: string;
  EndTime: string;
}

interface BackendSubmission {
  ID: string;
  FullName: string;
  Email: string;
  TopicTitle: string;
  Status?: string;
  CreatedAt?: string;
  ApprovedAt?: string | null;
  SlideFileName?: string | null;
}

@Injectable({
  providedIn: 'root'
})
export class ApplicationService {
  public readonly applications = signal<SpeakerApplication[]>([]);
  public readonly timelineEntries = signal<TimelineEntry[]>([]);
  public readonly isPresentationUploadEnabled = signal(false);
  
  private readonly apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) {
    this.loadSettings();
  }

  // ---------------------------------------------------------
  // Public APIs
  // ---------------------------------------------------------

  public submitApplication(application: { applicantName: string, email: string, topic: string }): void {
    const req = {
      full_name: application.applicantName,
      email: application.email,
      topic_title: application.topic
    };

    this.http.post(`${this.apiUrl}/submissions`, req).subscribe({
      next: () => {
        // Option to reload if needed, but typically users don't see pending from home
      },
      error: (err) => devLog('Failed to submit application', err)
    });
  }

  public uploadSlide(speakerId: string, file: File): Promise<void> {
    const formData = new FormData();
    formData.append('file', file);
    
    return new Promise((resolve, reject) => {
      this.http.put(`${this.apiUrl}/submissions/${speakerId}/slide`, formData).subscribe({
        next: () => resolve(),
        error: (err) => {
          devLog('Failed to upload slide', err);
          reject(err);
        }
      });
    });
  }

  public readonly publicApprovedSpeakers = signal<PublicSpeaker[]>([]);

  public loadAppConfig(): Observable<boolean> {
    return this.http.get<{ schedule_polling_enabled: boolean }>(`${this.apiUrl}/config`).pipe(
      map(res => res.schedule_polling_enabled ?? false),
      catchError(() => of(false))
    );
  }

  public loadTimeline(): void {
    this.http.get<{data: BackendSchedule[]}>(`${this.apiUrl}/schedule`).subscribe({
      next: (res) => {
        const mapped = res.data.map(entry => ({
          id: entry.ID,
          applicationId: entry.SubmissionID ?? undefined,
          speaker: entry.Submission?.FullName || 'Break / Event',
          topic: entry.Submission?.TopicTitle || entry.ScheduleType || 'Break / Event',
          startTime: entry.StartTime,
          endTime: entry.EndTime
        }));
        this.timelineEntries.set(mapped);
      },
      error: (err) => devLog('Failed to load timeline', err)
    });
  }

  public loadPublicApprovedSpeakers(): void {
    this.http.get<{data: PublicSpeaker[]}>(`${this.apiUrl}/speakers`).subscribe({
      next: (res) => {
        this.publicApprovedSpeakers.set(res.data || []);
      },
      error: (err) => devLog('Failed to load public speakers', err)
    });
  }

  public loadSettings(): void {
    this.http.get<{ presentation_upload_enabled: boolean }>(`${this.apiUrl}/settings`).subscribe({
      next: (res) => {
        this.isPresentationUploadEnabled.set(res.presentation_upload_enabled ?? false);
      },
      error: (err) => devLog('Failed to load settings', err)
    });
  }

  public setPresentationUploadEnabled(isEnabled: boolean): void {
    // Optimistic update
    this.isPresentationUploadEnabled.set(isEnabled);
    
    this.http.put(`${this.apiUrl}/admin/settings/presentation_upload_enabled`, { enabled: isEnabled }).subscribe({
      next: () => devLog('Presentation upload setting updated successfully'),
      error: (err) => {
        devLog('Failed to update presentation upload setting', err);
        // Revert on failure
        this.isPresentationUploadEnabled.set(!isEnabled);
      }
    });
  }

  // ---------------------------------------------------------
  // Admin APIs (Require JWT)
  // ---------------------------------------------------------

  public loadAdminData(): void {
    this.loadPending();
    this.loadApproved();
    this.loadTimeline();
  }

  public loadPending(): void {
    this.http.get<{data: BackendSubmission[]}>(`${this.apiUrl}/admin/submissions/pending`).subscribe({
      next: (res) => {
        const mapped = res.data.map(this.mapBackendApplication);
        this.applications.update(apps => {
          const others = apps.filter(a => a.status !== 'pending');
          return [...others, ...mapped];
        });
      },
      error: (err) => devLog('Failed to load pending', err)
    });
  }

  public loadApproved(): void {
    this.http.get<{data: BackendSubmission[]}>(`${this.apiUrl}/admin/submissions/approved`).subscribe({
      next: (res) => {
        const mapped = res.data.map(this.mapBackendApplication);
        this.applications.update(apps => {
          const others = apps.filter(a => a.status !== 'approved');
          return [...others, ...mapped];
        });
      },
      error: (err) => devLog('Failed to load approved', err)
    });
  }

  public approveApplication(applicationId: string): Promise<void> {
    return new Promise((resolve, reject) => {
      this.http.put(`${this.apiUrl}/admin/submissions/${applicationId}/approve`, {}).subscribe({
        next: () => {
          this.loadAdminData();
          resolve();
        },
        error: (err) => {
          devLog('Failed to approve application', err);
          reject(err);
        }
      });
    });
  }

  public rejectApplication(applicationId: string): Promise<void> {
    return new Promise((resolve, reject) => {
      this.http.delete(`${this.apiUrl}/admin/submissions/${applicationId}`).subscribe({
        next: () => {
          this.loadAdminData();
          resolve();
        },
        error: (err) => {
          devLog('Failed to reject application', err);
          reject(err);
        }
      });
    });
  }

  public updateSubmissionInfo(applicationId: string, fullName: string, topicTitle: string): Promise<void> {
    return new Promise((resolve, reject) => {
      this.http.put(`${this.apiUrl}/admin/submissions/${applicationId}/info`, {
        full_name: fullName,
        topic_title: topicTitle,
      }).subscribe({
        next: () => {
          this.loadAdminData();
          resolve();
        },
        error: (err) => {
          devLog('Failed to update submission info', err);
          reject(err);
        }
      });
    });
  }

  public downloadSlide(applicationId: string): void {
    this.http.get(`${this.apiUrl}/admin/submissions/${applicationId}/slide`, { responseType: 'blob', observe: 'response' }).subscribe({
      next: (response) => {
        const blob = response.body as Blob;
        let fileName = 'slide';
        
        const contentDisposition = response.headers.get('Content-Disposition');
        if (contentDisposition) {
          // Try filename* first (RFC 5987/6266) for UTF-8 support
          const fileNameStarMatch = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i);
          if (fileNameStarMatch && fileNameStarMatch.length === 2) {
            fileName = decodeURIComponent(fileNameStarMatch[1]);
          } else {
            // Fallback to standard filename
            const fileNameMatch = contentDisposition.match(/filename="?([^";]+)"?/);
            if (fileNameMatch && fileNameMatch.length === 2) {
              fileName = fileNameMatch[1];
            }
          }
        }

        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = fileName;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
      },
      error: () => alert('Failed to download slide or slide not found.')
    });
  }

  public downloadZip(applicationIds: string[]): void {
    this.http.post(`${this.apiUrl}/admin/submissions/slides/zip`, { ids: applicationIds }, { responseType: 'blob' }).subscribe({
      next: (blob) => {
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'slides.zip';
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
      },
      error: () => alert('Failed to download ZIP file.')
    });
  }

  public saveTimelineEntry(
    entry: { applicationId?: string, startTime: string, endTime: string, id?: string },
    isEdit: boolean = false,
  ): Promise<void> {
    const req = {
      submission_id: entry.applicationId,
      schedule_type: entry.applicationId ? 'SPEAKER' : 'BREAK',
      start_time: entry.startTime,
      end_time: entry.endTime
    };

    return new Promise((resolve, reject) => {
      if (isEdit && entry.id) {
        this.http.delete(`${this.apiUrl}/admin/schedule/${entry.id}`).subscribe({
          next: () => {
            this.http.post(`${this.apiUrl}/admin/schedule`, req).subscribe({
              next: () => {
                this.loadTimeline();
                resolve();
              },
              error: (err) => {
                devLog('Failed to add updated schedule slot', err);
                reject(err);
              }
            });
          },
          error: (err) => {
            devLog('Failed to delete old schedule slot for update', err);
            reject(err);
          }
        });
        return;
      }

      this.http.post(`${this.apiUrl}/admin/schedule`, req).subscribe({
        next: () => {
          this.loadTimeline();
          resolve();
        },
        error: (err) => {
          devLog('Failed to save schedule slot', err);
          reject(err);
        }
      });
    });
  }

  public deleteTimelineEntry(entryId: string): Promise<void> {
    return new Promise((resolve, reject) => {
      this.http.delete(`${this.apiUrl}/admin/schedule/${entryId}`).subscribe({
        next: () => {
          this.loadTimeline();
          resolve();
        },
        error: (err) => {
          devLog('Failed to delete schedule slot', err);
          reject(err);
        }
      });
    });
  }

  // Helper
  private mapBackendApplication(b: BackendSubmission): SpeakerApplication {
    return {
      id: b.ID,
      applicantName: b.FullName,
      email: b.Email,
      topic: b.TopicTitle,
      status: b.Status?.toLowerCase() as ApplicationStatus,
      submissionDate: b.CreatedAt ? new Date(b.CreatedAt).toISOString().split('T')[0] : '',
      approvedAt: b.ApprovedAt ? new Date(b.ApprovedAt).toISOString().split('T')[0] : undefined,
      slideFileName: b.SlideFileName ?? undefined,
    };
  }

}

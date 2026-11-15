import { CommonModule } from '@angular/common';
import { Component, DestroyRef, computed, inject, signal, OnInit, AfterViewInit, ViewChildren, QueryList, ElementRef, effect } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ButtonComponent } from '../../components/button/button.component';
import { ApplicationService, TimelineEntry } from '../../services/application.service';
import { PresentationUploadData, UploadPresentationComponent } from '../../components/upload-presentation/upload-presentation.component';

interface HomeFeature {
  icon: string;
  title: string;
  description: string;
  tone: 'blue' | 'yellow' | 'violet';
}

interface QueueDisplayEntry extends TimelineEntry {
  status: 'past' | 'live' | 'upcoming';
}

interface TimelineItem {
  icon: string;
  side: 'left' | 'right';
  date: string;
  title: string;
  description?: string;
}

@Component({
  selector: 'app-home-page',
  imports: [CommonModule, FormsModule, ButtonComponent, UploadPresentationComponent],
  templateUrl: './home.page.html',
  styleUrl: './home.page.scss',
})
export class HomePage implements OnInit, AfterViewInit {
  private readonly destroyRef = inject(DestroyRef);
  protected readonly applicationService = inject(ApplicationService);

  protected readonly isSubmissionModalOpen = signal(false);
  protected readonly isSuccessModalOpen = signal(false);
  protected readonly isUploadSuccessModalOpen = signal(false);
  protected readonly isUploadModalOpen = signal(false);
  protected readonly isHomeNavOpen = signal(false);
  protected readonly emailFormatError = signal('');
  
  protected readonly queueScrollThreshold = 10;
  protected readonly queueRowHeight = 90;
  protected readonly currentMinutes = signal(this.getCurrentMinutes());
  protected activeSection = signal('about');
  protected currentTheme = signal<'light' | 'dark'>('light');
  
  @ViewChildren('navLink') navLinkElements!: QueryList<ElementRef>;
  protected indicatorStyle = signal({ left: 0, width: 0, opacity: 0 });

  protected readonly navLinks = [
    { label: 'About', id: 'about' },
    { label: 'Timeline', id: 'timeline' },
    { label: 'Speakers', id: 'speakers' },
  ];
  protected readonly features: HomeFeature[] = [
    {
      icon: 'groups',
      title: 'Confidence',
      description: 'Build the courage to speak in front of others with clarity, calm, and self-assurance.',
      tone: 'blue',
    },
    {
      icon: 'record_voice_over',
      title: 'Practice',
      description: 'Gain real speaking experience in a supportive environment where every session is a chance to improve.',
      tone: 'yellow',
    },
    {
      icon: 'trending_up',
      title: 'Growth',
      description: 'Strengthen your communication skills step by step through speaking, listening, and feedback.',
      tone: 'violet',
    },
  ];

  protected submissionForm = {
    applicantName: '',
    email: '',
    topic: ''
  };

  ngOnInit(): void {
    const saved = localStorage.getItem('tex-theme') as 'light' | 'dark' | null;
    if (saved === 'light' || saved === 'dark') {
      this.currentTheme.set(saved);
    } else {
      this.currentTheme.set('light');
    }
  }

  protected toggleTheme(): void {
    const next = this.currentTheme() === 'light' ? 'dark' : 'light';
    this.currentTheme.set(next);
    localStorage.setItem('tex-theme', next);
  }

  protected toggleHomeNav(): void {
    this.isHomeNavOpen.update((isOpen) => !isOpen);
  }

  protected closeHomeNav(): void {
    this.isHomeNavOpen.set(false);
  }

  ngAfterViewInit(): void {
    const observerOptions = {
      root: null,
      rootMargin: '-40% 0px -60% 0px',
      threshold: 0
    };

    const observer = new IntersectionObserver((entries) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          const id = entry.target.getAttribute('id');
          if (id && this.navLinks.some(link => link.id === id)) {
            this.activeSection.set(id);
          }
        }
      });
    }, observerOptions);

    this.navLinks.forEach(link => {
      const element = document.getElementById(link.id);
      if (element) {
        observer.observe(element);
      }
    });

    this.destroyRef.onDestroy(() => observer.disconnect());
  }

  protected readonly queueEntries = computed<QueueDisplayEntry[]>(() => {
    const now = this.currentMinutes();
    
    const timeline = this.applicationService.timelineEntries();
    const publicSpeakers = this.applicationService.publicApprovedSpeakers();
    
    // Map the scheduled entries
    const scheduledIds = new Set<string>();
    const mappedSchedule = timeline.map(entry => {
      if (entry.applicationId) scheduledIds.add(entry.applicationId);
      return {
        ...entry,
        status: this.getQueueStatus(entry, now),
      };
    });

    // Add unscheduled approved speakers
    const unscheduledSpeakers = publicSpeakers
      .filter(speaker => !scheduledIds.has(speaker.id))
      .map(speaker => {
        const entry: TimelineEntry = {
          id: speaker.id,
          applicationId: speaker.id,
          speaker: speaker.full_name,
          topic: speaker.topic_title,
          startTime: 'TBA',
          endTime: 'TBA'
        };
        return {
          ...entry,
          status: this.getQueueStatus(entry, now)
        };
      });

    return [...mappedSchedule, ...unscheduledSpeakers];
  });

  protected readonly timelineItems: TimelineItem[] = [
    { icon: 'campaign', side: 'left', date: '6 May 2026', title: 'Start Submit Topic' },
    { icon: 'event_busy', side: 'right', date: '22 May 2026', title: 'End Submit Topic' },
    { icon: 'calendar_month', side: 'left', date: '29 May 2026', title: 'Announce Speaking Schedule' },
    { icon: 'present_to_all', side: 'right', date: '29 May - 10 Jun 2026', title: 'Submit Presentation Slide' },
    { icon: 'celebration', side: 'left', date: '13 June 2026', title: 'TEX Talk Day', description: 'Location : To be announced' },
  ];

  protected get isQueueScrollable(): boolean {
    return this.applicationService.timelineEntries().length > this.queueScrollThreshold;
  }

  protected get queueListMaxHeight(): number | null {
    return this.isQueueScrollable ? this.queueScrollThreshold * this.queueRowHeight : null;
  }

  protected openSubmissionModal(): void {
    this.closeHomeNav();
    this.submissionForm = { applicantName: '', email: '', topic: '' };
    this.emailFormatError.set('');
    this.isSubmissionModalOpen.set(true);
  }

  protected closeSubmissionModal(): void {
    this.isSubmissionModalOpen.set(false);
    this.emailFormatError.set('');
  }

  protected closeSuccessModal(): void {
    this.isSuccessModalOpen.set(false);
  }

  protected openUploadModal(): void {
    if (!this.applicationService.isPresentationUploadEnabled()) {
      return;
    }

    this.isUploadModalOpen.set(true);
  }

  protected closeUploadModal(): void {
    this.isUploadModalOpen.set(false);
  }

  protected closeUploadSuccessModal(): void {
    this.isUploadSuccessModalOpen.set(false);
  }

  protected handleUploadSubmit(data: PresentationUploadData): void {
    this.applicationService.uploadSlide(data.speakerId, data.file).then(() => {
      this.closeUploadModal();
      this.isUploadSuccessModalOpen.set(true);
    }).catch(() => {
      alert('Failed to upload slide. Please try again.');
    });
  }

  protected submitForm(): void {
    if (!this.submissionForm.applicantName || !this.submissionForm.email || !this.submissionForm.topic) {
      alert('Please fill out all fields');
      return;
    }

    if (!this.isValidEmail(this.submissionForm.email)) {
      this.emailFormatError.set('Please enter a valid email address.');
      return;
    }
    
    this.applicationService.submitApplication({
      applicantName: this.submissionForm.applicantName,
      email: this.submissionForm.email,
      topic: this.submissionForm.topic
    });
    
    this.closeSubmissionModal();
    this.isSuccessModalOpen.set(true);
  }

  protected validateEmailFormat(): void {
    const email = this.submissionForm.email.trim();
    this.emailFormatError.set(
      email && !this.isValidEmail(email) ? 'Please enter a valid email address.' : '',
    );
  }

  protected scrollToSpeakers(): void {
    this.closeHomeNav();
    const element = document.getElementById('speakers');
    if (element) {
      element.scrollIntoView({ behavior: 'smooth' });
    }
  }

  constructor() {
    effect(() => {
      this.activeSection();
      setTimeout(() => this.updateIndicator(), 10);
    });

    const clockTimer = window.setInterval(() => {
      this.currentMinutes.set(this.getCurrentMinutes());
    }, 60_000);

    let dataTimer: number | undefined;
    this.applicationService.loadAppConfig().subscribe(pollingEnabled => {
      if (pollingEnabled) {
        this.applicationService.loadTimeline();
        this.applicationService.loadPublicApprovedSpeakers();
        dataTimer = window.setInterval(() => {
          this.applicationService.loadTimeline();
          this.applicationService.loadPublicApprovedSpeakers();
        }, 5 * 60_000);
      }
    });

    const onResize = () => this.updateIndicator();
    window.addEventListener('resize', onResize);

    this.destroyRef.onDestroy(() => {
      window.clearInterval(clockTimer);
      if (dataTimer !== undefined) window.clearInterval(dataTimer);
      window.removeEventListener('resize', onResize);
    });
  }

  private getQueueStatus(entry: TimelineEntry, currentMinutes: number): QueueDisplayEntry['status'] {
    if (entry.startTime === 'TBA') {
      return 'upcoming';
    }

    const start = this.timeToMinutes(entry.startTime);
    const end = this.timeToMinutes(entry.endTime);

    if (currentMinutes >= start && currentMinutes <= end) {
      return 'live';
    }

    if (currentMinutes < start) {
      return 'upcoming';
    }

    return 'past';
  }

  private getCurrentMinutes(): number {
    const now = new Date();
    return now.getHours() * 60 + now.getMinutes();
  }

  private timeToMinutes(time: string): number {
    const [hours, minutes] = time.split(':').map(Number);
    return hours * 60 + minutes;
  }

  private updateIndicator(): void {
    if (!this.navLinkElements) return;

    const activeIndex = this.navLinks.findIndex(l => l.id === this.activeSection());
    if (activeIndex >= 0) {
      const elArray = this.navLinkElements.toArray();
      if (elArray[activeIndex]) {
        const el = elArray[activeIndex].nativeElement;
        this.indicatorStyle.set({
          left: el.offsetLeft,
          width: el.offsetWidth,
          opacity: 1
        });
      }
    }
  }

  private isValidEmail(email: string): boolean {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim());
  }
}

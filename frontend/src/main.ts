import { bootstrapApplication } from '@angular/platform-browser';

import { appConfig } from './app/app.config';
import { AppComponent } from './app/app.component';
import { environment } from './environments/environment';

if (environment.production) {
  const noop = () => {};
  window.console.log = noop;
  window.console.warn = noop;
  window.console.info = noop;
  window.console.debug = noop;
  window.console.error = noop;
}

bootstrapApplication(AppComponent, appConfig).catch((error: unknown) => console.error(error));

/**
 * Sovereign Settings - Authenticated Preference Management
 */

export interface UserSettings {
  theme: 'dark' | 'light';
  notifications: boolean;
  modelTier: 'prime' | 'mist' | 'phi';
  retentionDays: number;
}

export class SettingsManager {
  private currentSettings: UserSettings | null = null;
  private debounceTimer: number | null = null;

  constructor() {
    this.load();
  }

  public async load(): Promise<UserSettings | null> {
    try {
      const response = await fetch('/api/settings');
      if (!response.ok) throw new Error('Failed to load settings');
      
      this.currentSettings = await response.json();
      return this.currentSettings;
    } catch (err) {
      console.error('[SETTINGS] Load error:', err);
      return null;
    }
  }

  public update<K extends keyof UserSettings>(key: K, value: UserSettings[K]): void {
    if (!this.currentSettings) return;

    // Type-safe update
    this.currentSettings[key] = value;

    // Debounce save by 500ms
    if (this.debounceTimer) clearTimeout(this.debounceTimer);
    this.debounceTimer = window.setTimeout(() => this.save(), 500);
  }

  private async save(): Promise<void> {
    if (!this.currentSettings) return;

    // Validation
    if (this.currentSettings.retentionDays < 0) {
      console.error('[SETTINGS] Invalid retention days');
      return;
    }

    try {
      const response = await fetch('/api/settings', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(this.currentSettings),
      });

      if (!response.ok) throw new Error('Save failed');
      console.log('[SETTINGS] Preferences saved successfully');
    } catch (err) {
      console.error('[SETTINGS] Save error:', err);
    }
  }

  public get<K extends keyof UserSettings>(key: K): UserSettings[K] | undefined {
    return this.currentSettings?.[key];
  }
}

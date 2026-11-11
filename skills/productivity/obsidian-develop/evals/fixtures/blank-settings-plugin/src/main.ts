import { Plugin } from "obsidian";
import { DEFAULT_SETTINGS, LinkStatsSettings, LinkStatsSettingTab } from "./settings";

export default class LinkStatsPlugin extends Plugin {
	settings: LinkStatsSettings;

	async onload() {
		this.addSettingTab(new LinkStatsSettingTab(this.app, this));
		this.loadSettings();
		this.addRibbonIcon("link", "Show link stats", () => this.showStats());
	}

	async loadSettings() {
		this.settings = Object.assign({}, DEFAULT_SETTINGS, await this.loadData());
	}

	async saveSettings() {
		await this.saveData(this.settings);
	}

	showStats() {
		const file = this.app.workspace.getActiveFile();
		if (!file) return;
		const links = this.app.metadataCache.getFileCache(file)?.links ?? [];
		console.log(`${file.basename}: ${links.length} links`);
	}
}

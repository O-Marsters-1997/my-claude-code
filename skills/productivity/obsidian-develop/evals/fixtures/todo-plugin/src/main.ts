import { Plugin } from "obsidian";
import { DEFAULT_SETTINGS, TodoHelperSettings, TodoHelperSettingTab } from "./settings";

export default class TodoHelperPlugin extends Plugin {
	settings: TodoHelperSettings;

	async onload() {
		await this.loadSettings();
		this.addSettingTab(new TodoHelperSettingTab(this.app, this));
	}

	async loadSettings() {
		this.settings = Object.assign({}, DEFAULT_SETTINGS, await this.loadData());
	}

	async saveSettings() {
		await this.saveData(this.settings);
	}
}

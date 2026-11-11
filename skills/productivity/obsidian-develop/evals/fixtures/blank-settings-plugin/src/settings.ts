import { App, PluginSettingTab, Setting } from "obsidian";
import type LinkStatsPlugin from "./main";

export interface LinkStatsSettings {
	countEmbeds: boolean;
	excludedFolders: string[];
}

export const DEFAULT_SETTINGS: LinkStatsSettings = {
	countEmbeds: true,
	excludedFolders: [],
};

export class LinkStatsSettingTab extends PluginSettingTab {
	plugin: LinkStatsPlugin;

	constructor(app: App, plugin: LinkStatsPlugin) {
		super(app, plugin);
		this.plugin = plugin;
	}

	display(): void {
		const { containerEl } = this;
		containerEl.empty();
		const section = createDiv({ cls: "link-stats-settings" });

		new Setting(section)
			.setName("Count embeds")
			.addToggle((t) =>
				t.setValue(this.plugin.settings.countEmbeds).onChange(async (v) => {
					this.plugin.settings.countEmbeds = v;
					await this.plugin.saveSettings();
				}),
			);

		new Setting(section)
			.setName("Excluded folders")
			.setDesc("Comma-separated.")
			.addText((t) =>
				t.setValue(this.plugin.settings.excludedFolders.join(", ")).onChange(async (v) => {
					this.plugin.settings.excludedFolders = v.split(",").map((s) => s.trim());
					await this.plugin.saveSettings();
				}),
			);
	}
}

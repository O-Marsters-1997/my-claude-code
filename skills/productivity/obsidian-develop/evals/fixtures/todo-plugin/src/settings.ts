import { App, PluginSettingTab, Setting } from "obsidian";
import type TodoHelperPlugin from "./main";

export interface TodoHelperSettings {
	dueDateField: string;
}

export const DEFAULT_SETTINGS: TodoHelperSettings = {
	dueDateField: "due",
};

export class TodoHelperSettingTab extends PluginSettingTab {
	plugin: TodoHelperPlugin;

	constructor(app: App, plugin: TodoHelperPlugin) {
		super(app, plugin);
		this.plugin = plugin;
	}

	display(): void {
		const { containerEl } = this;
		containerEl.empty();
		new Setting(containerEl)
			.setName("Due date field")
			.setDesc("Frontmatter key that holds the due date.")
			.addText((text) =>
				text.setValue(this.plugin.settings.dueDateField).onChange(async (value) => {
					this.plugin.settings.dueDateField = value;
					await this.plugin.saveSettings();
				}),
			);
	}
}

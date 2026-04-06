#!/usr/bin/env node

import { spawn } from "node:child_process";
import { parseArgs } from "node:util";

const commands = {
	help: {
		description: "Show help",
		run: () => {
			const maxLen = Math.max(...Object.keys(commands).map((k) => k.length));
			console.log("Usage: node dev.mjs <command>\n");
			for (const [name, cmd] of Object.entries(commands)) {
				console.log(`  ${name.padEnd(maxLen + 2)} ${cmd.description}`);
			}
		},
	},
	setup: {
		description: "Setup the project",
		run: () =>
			sequential([
				{ cmd: "npm", args: ["install"], cwd: "frontend", label: "frontend" },
				{
					cmd: "go",
					args: ["mod", "download"],
					cwd: "api",
					label: "api",
				},
			]),
	},
	lint: {
		description: "Run linters",
		run: () =>
			sequential([
				{ cmd: "golangci-lint", args: ["run"], cwd: "api", label: "golangci-lint" },
				{
					cmd: "npm",
					args: ["run", "lint"],
					cwd: "frontend",
					label: "oxlint",
				},
				{
					cmd: "npm",
					args: ["run", "format"],
					cwd: "frontend",
					label: "oxfmt",
				},
				{
					cmd: "npm",
					args: ["run", "tsc"],
					cwd: "frontend",
					label: "vue-tsc",
				},
				{
					cmd: "npm",
					args: ["run", "test:run"],
					cwd: "frontend",
					label: "vitest",
				},
			]),
	},
	"lint-fix": {
		description: "Fix lint issues",
		run: () =>
			sequential([
				{ cmd: "npm", args: ["run", "lint:fix"], cwd: "frontend", label: "oxlint" },
				{ cmd: "npm", args: ["run", "format:fix"], cwd: "frontend", label: "oxfmt" },
			]),
	},
	generate: {
		description: "Generate API code (sqlc + oapi-codegen)",
		run: () =>
			sequential([{ cmd: "make", args: ["generate"], cwd: "api", label: "generate" }]),
	},
	migrate: {
		description: "Run database migrations",
		run: () =>
			sequential([
				{ cmd: "go", args: ["run", ".", "migrate", "up"], cwd: "api", label: "migrate" },
			]),
	},
	"migrate-status": {
		description: "Show migration status",
		run: () =>
			sequential([
				{
					cmd: "go",
					args: ["run", ".", "migrate", "status"],
					cwd: "api",
					label: "migrate-status",
				},
			]),
	},
	"drop-db": {
		description: "Drop the PostgreSQL database",
		run: () =>
			sequential([
				{
					cmd: "docker",
					args: [
						"compose",
						"exec",
						"-T",
						"db",
						"psql",
						"-U",
						"shopmon",
						"-c",
						"DROP SCHEMA IF EXISTS drizzle CASCADE; DROP SCHEMA public CASCADE; CREATE SCHEMA public;",
					],
					label: "drop-db",
				},
			]),
	},
	"load-fixtures": {
		description: "Drop DB, migrate, and seed fixtures",
		run: async () => {
			await commands["drop-db"].run();
			await commands["migrate"].run();
			await sequential([
				{
					cmd: "go",
					args: ["run", ".", "fixtures"],
					cwd: "api",
					label: "fixtures",
				},
			]);
		},
	},
	test: {
		description: "Run tests",
		run: () =>
			sequential([
				{
					cmd: "go",
					args: ["test", "./internal/...", "-count=1", "-timeout", "120s"],
					cwd: "api",
					label: "test",
				},
			]),
	},
	build: {
		description: "Build the API binary",
		run: () =>
			sequential([
				{
					cmd: "go",
					args: ["build", "-o", "bin/shopmon", "."],
					cwd: "api",
					label: "build",
				},
			]),
	},
	dev: {
		description: "Run API, worker, and frontend in dev mode (parallel)",
		run: () =>
			parallel([
				{
					cmd: "air",
					args: [],
					cwd: "api",
					label: "api",
					env: { LISTEN_ADDR: "127.0.0.1:5789" },
				},
				{
					cmd: "air",
					args: ["-c", ".air-worker.toml"],
					cwd: "api",
					label: "worker",
				},
				{
					cmd: "npm",
					args: ["run", "dev"],
					cwd: "frontend",
					label: "frontend",
				},
			]),
	},
	"dev-worker": {
		description: "Run background worker",
		run: () =>
			sequential([
				{
					cmd: "go",
					args: ["run", ".", "worker"],
					cwd: "api",
					label: "worker",
				},
			]),
	},
	up: {
		description: "Start infrastructure (DB, Redis, demo shop)",
		run: async () => {
			await sequential([
				{ cmd: "docker", args: ["compose", "up", "-d"], label: "docker" },
			]);
			console.log("\nDemo shop: http://localhost:3889");
			console.log("Mailpit:   http://localhost:8025");
		},
	},
	stop: {
		description: "Stop infrastructure",
		run: () =>
			sequential([
				{ cmd: "docker", args: ["compose", "stop"], label: "docker" },
			]),
	},
};

// --- helpers ---

const COLORS = {
	reset: "\x1b[0m",
	bold: "\x1b[1m",
	dim: "\x1b[2m",
	labels: [
		"\x1b[36m", // cyan
		"\x1b[35m", // magenta
		"\x1b[33m", // yellow
		"\x1b[32m", // green
		"\x1b[34m", // blue
	],
};

let colorIdx = 0;
function nextColor() {
	return COLORS.labels[colorIdx++ % COLORS.labels.length];
}

function prefixStream(stream, prefix) {
	let buffer = "";
	stream.on("data", (chunk) => {
		buffer += chunk.toString();
		const lines = buffer.split("\n");
		buffer = lines.pop(); // keep incomplete line in buffer
		for (const line of lines) {
			process.stdout.write(`${prefix} ${line}\n`);
		}
	});
	stream.on("end", () => {
		if (buffer.length > 0) {
			process.stdout.write(`${prefix} ${buffer}\n`);
		}
	});
}

function run({ cmd, args = [], cwd, label, env = {} }) {
	return new Promise((resolve, reject) => {
		const color = nextColor();
		const prefix = `${color}${COLORS.bold}[${label}]${COLORS.reset}`;
		const proc = spawn(cmd, args, {
			cwd: cwd ? new URL(cwd, import.meta.url).pathname : undefined,
			env: { ...process.env, ...env },
			stdio: ["ignore", "pipe", "pipe"],
		});

		prefixStream(proc.stdout, prefix);
		prefixStream(proc.stderr, prefix);

		proc.on("close", (code) => {
			if (code !== 0) {
				reject(new Error(`[${label}] exited with code ${code}`));
			} else {
				resolve();
			}
		});
		proc.on("error", reject);
	});
}

async function sequential(tasks) {
	for (const task of tasks) {
		await run(task);
	}
}

function parallel(tasks) {
	// Forward SIGINT/SIGTERM to children for clean shutdown
	const children = [];

	const promises = tasks.map((task) => {
		return new Promise((resolve, reject) => {
			const color = nextColor();
			const prefix = `${color}${COLORS.bold}[${task.label}]${COLORS.reset}`;
			const proc = spawn(task.cmd, task.args || [], {
				cwd: task.cwd ? new URL(task.cwd, import.meta.url).pathname : undefined,
				env: { ...process.env, ...task.env },
				stdio: ["ignore", "pipe", "pipe"],
			});

			children.push(proc);
			prefixStream(proc.stdout, prefix);
			prefixStream(proc.stderr, prefix);

			proc.on("close", (code) => {
				if (code !== 0 && code !== null) {
					reject(new Error(`[${task.label}] exited with code ${code}`));
				} else {
					resolve();
				}
			});
			proc.on("error", reject);
		});
	});

	const cleanup = (signal) => {
		for (const child of children) {
			child.kill(signal);
		}
	};

	process.on("SIGINT", () => cleanup("SIGINT"));
	process.on("SIGTERM", () => cleanup("SIGTERM"));

	return Promise.all(promises);
}

// --- main ---

const { positionals } = parseArgs({ allowPositionals: true });
const command = positionals[0] || "help";

if (!commands[command]) {
	console.error(`Unknown command: ${command}`);
	commands.help.run();
	process.exit(1);
}

try {
	await commands[command].run();
} catch (err) {
	console.error(`\n${COLORS.bold}Error:${COLORS.reset} ${err.message}`);
	process.exit(1);
}

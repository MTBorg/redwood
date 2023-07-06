use std::path::PathBuf;

use clap::{ArgEnum, Parser, Subcommand};

#[derive(Debug, Parser)]
#[clap(name = "redwood")]
pub struct Cli {
    /// The path of the log file to log to.
    /// Defaults to $HOME/.redwood.log
    #[clap(long)]
    pub log_file_path: Option<PathBuf>,

    /// The level to use when logging to file.
    /// Defaults to off.
    #[clap(long, arg_enum)]
    pub log_file_level: Option<LogFileLevel>,

    #[clap(subcommand)]
    pub command: Commands,
}

#[derive(Clone, Debug, ArgEnum)]
pub enum LogFileLevel {
    Off,
    Error,
    Warn,
    Info,
    Debug,
    Trace,
}

#[derive(Debug, Subcommand)]
pub enum Commands {
    /// Create new worktree
    #[clap(arg_required_else_help = true)]
    New {
        #[clap(required = true)]
        worktree_name: String,
        #[clap(required = false, parse(from_os_str))]
        repo_path: Option<PathBuf>,
        #[clap(long)]
        tmux_session_name: Option<String>,
    },
    /// Open existing worktree configuration
    Open {
        #[clap(required = true)]
        identifier: String,
    },
    /// Delete worktree configuration
    Delete {
        #[clap(required = true)]
        identifier: String,
    },
    /// Import existing worktree
    Import {
        #[clap(required = true, parse(from_os_str))]
        worktree_path: PathBuf,
        #[clap(long)]
        tmux_session_name: Option<String>,
    },
    /// List existing worktree configurations
    List {},
    /// Print version of Redwood
    Version {},
}

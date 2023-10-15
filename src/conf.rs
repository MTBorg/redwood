use serde::{Deserialize, Serialize};
use std::env;
use std::path::{Path, PathBuf};

use crate::error::RedwoodError::*;
use crate::Result;

const CONFIG_DIRECTORY_NAME: &str = "redwood";
const CONFIG_FILE_NAME: &str = "conf.json";

#[derive(Serialize, Deserialize, Debug, Default)]
#[serde(rename_all = "camelCase")]
pub struct Config {
    worktrees: Vec<WorktreeConfig>,
}

pub trait ConfigWriter {
    fn write(&self, cfg: &Config) -> Result<()>;
}

struct ConfigWriterImpl {
    config_path: PathBuf,
}

pub fn new_writer(config_path: &Path) -> impl ConfigWriter {
    ConfigWriterImpl {
        config_path: config_path.to_owned(),
    }
}

impl ConfigWriter for ConfigWriterImpl {
    fn write(&self, cfg: &Config) -> Result<()> {
        let contents = cfg.serialize()?;

        // Make sure that the directory exists before writing to it
        let config_dir = get_config_dir()?;
        if let Err(e) = std::fs::create_dir_all(config_dir) {
            return Err(ConfigWriteError(e.to_string()));
        }

        let config_path = &self.config_path;

        match std::fs::write(config_path, contents) {
            Ok(()) => Ok(()),
            Err(err) => Err(ConfigWriteError(err.to_string())),
        }
    }
}

impl Config {
    pub fn add_worktree(&mut self, wt: WorktreeConfig) -> Result<()> {
        if self
            .worktrees
            .iter()
            .any(|wt2| wt2.repo_path == wt.repo_path && wt2.worktree_name == wt.worktree_name)
        {
            return Err(WorkTreeConfigAlreadyExists);
        }
        self.worktrees.push(wt);
        Ok(())
    }

    pub fn remove_worktree(&mut self, worktree_name: &str) -> Result<()> {
        let (cfg_index, _) = match self.find(worktree_name) {
            Some(index) => index,
            None => {
                return Err(WorkTreeConfigNotFound {
                    worktree_name: String::from(worktree_name),
                })
            }
        };
        self.worktrees.remove(cfg_index);
        Ok(())
    }

    pub fn new() -> Self {
        Config { worktrees: vec![] }
    }

    pub fn serialize(&self) -> Result<String> {
        match serde_json::to_string_pretty(&self) {
            Ok(contents) => Ok(contents),
            Err(msg) => Err(ConfigWriteError(msg.to_string())), // TODO: More appropriate
                                                                // error
        }
    }

    pub fn worktrees(&self) -> &Vec<WorktreeConfig> {
        &self.worktrees
    }

    pub fn find(&self, identifier: &str) -> Option<(usize, &WorktreeConfig)> {
        self.worktrees
            .iter()
            .enumerate()
            .find(|(_, wt)| identifier == wt.repo_path() || identifier == wt.worktree_name())
    }

    pub fn list(&self) -> Vec<WorktreeConfig> {
        let mut worktrees = self.worktrees().to_vec();

        worktrees.sort_by(|wt1, wt2| {
            wt1.repo_path()
                .to_lowercase()
                .cmp(&wt2.repo_path().to_lowercase())
        });
        worktrees
    }
}

#[derive(Serialize, Deserialize, Debug, Clone)]
#[serde(rename_all = "camelCase")]
pub struct WorktreeConfig {
    repo_path: String,
    worktree_name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    tmux_session_name: Option<String>,
}

impl WorktreeConfig {
    pub fn new(repo_path: &Path, worktree_name: &str) -> Self {
        WorktreeConfig {
            repo_path: repo_path.to_string_lossy().into_owned(),
            worktree_name: String::from(worktree_name),
            tmux_session_name: None,
        }
    }

    pub fn repo_path(&self) -> &str {
        &self.repo_path
    }

    pub fn worktree_name(&self) -> &str {
        &self.worktree_name
    }

    pub fn tmux_session_name(&self) -> Option<&str> {
        self.tmux_session_name.as_deref()
    }

    pub fn set_tmux_session_name(&mut self, name: &str) {
        self.tmux_session_name = Some(name.to_owned())
    }
}

pub fn read_config(config_path: &Path) -> Result<Config> {
    let content = match std::fs::read_to_string(config_path) {
        Ok(content) => content,
        Err(e) => {
            return match e.kind() {
                std::io::ErrorKind::NotFound => Err(ConfigNotFound),
                _ => Err(ConfigReadError(e.to_string())),
            }
        }
    };

    let config: Config = match serde_json::from_str(&content) {
        Ok(cfg) => cfg,
        Err(msg) => panic!("deserialize config {}", msg),
    };

    Ok(config)
}

fn get_config_dir() -> Result<PathBuf> {
    if let Some(path) = env::var_os("XDG_CONFIG_HOME") {
        return Ok(PathBuf::from(path).join(CONFIG_DIRECTORY_NAME));
    }
    if let Some(path) = env::var_os("HOME") {
        return Ok(PathBuf::from(path)
            .join(".config")
            .join(CONFIG_DIRECTORY_NAME));
    }
    Err(ConfigPathUnresolvable)
}

pub fn get_config_path() -> Result<PathBuf> {
    let config_path = get_config_dir()?;
    Ok(config_path.join(CONFIG_FILE_NAME))
}

mod tests {
    #[test]
    fn list() {
        use super::*;
        use crate::conf::WorktreeConfig;

        let mut cfg = crate::conf::Config::new();

        let wts: Vec<WorktreeConfig> =
            vec!["b", "a/b/c", "a/b/c/d", "a/b", "a/b/a", "a/b/a/e", "a"]
                .iter()
                .map(|p| Path::new(p))
                .map(|p| WorktreeConfig::new(p, ""))
                .collect();

        for wt in wts {
            cfg.add_worktree(wt).unwrap();
        }

        let v = cfg.list();
        assert_eq!(v[0].repo_path(), "a");
        assert_eq!(v[1].repo_path(), "a/b");
        assert_eq!(v[2].repo_path(), "a/b/a");
        assert_eq!(v[3].repo_path(), "a/b/a/e");
        assert_eq!(v[4].repo_path(), "a/b/c");
        assert_eq!(v[5].repo_path(), "a/b/c/d");
        assert_eq!(v[6].repo_path(), "b");
    }
}

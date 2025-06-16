use serde::{Serialize, Deserialize};
use std::fs::read_to_string;
use std::collections::HashMap;
use std::error::Error;
use std::path::Path;


#[derive(Serialize, Deserialize, Debug)]
pub struct GadgetVariable {
    pub name: String,
    pub description: String,
    pub default: Option<String>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Gadget {
    pub name: String,
    pub command: String,
    pub description: String,
    pub variables: Vec<GadgetVariable>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct GadgetStore(pub HashMap<String, Gadget>);

fn file_path() -> Result<String, Box<dyn Error>> {
    if std::env::consts::OS == "linux" {
        let home_dir: String = std::env::var("HOME")?;
        return Ok(format!("{}/.config/GoGoGadget/gadgets.json", home_dir));
    }
    let local_app_data: String = std::env::var("LOCALAPPDATA")?;
    Ok(format!("{}/GoGoGadget/gadgets.json", local_app_data))
}

impl GadgetStore {
    pub fn new() -> Result<Self, Box<dyn Error>> {
        let gadget_file_path: String = file_path()?;
        let parent_dir = Path::new(&gadget_file_path)
            .parent()
            .map(|p| p.exists())
            .unwrap_or(false);
        if !parent_dir {
            return Err("Gadget config directory does not exist".into());
        }
        let data: String = read_to_string(&gadget_file_path)?;
        let parsed: GadgetStore = serde_json::from_str(&data)?;
        Ok(parsed)
    }

    pub fn get_gadget(&self, name: &str) -> Option<&Gadget> {
        self.0.get(name)
    }

    pub fn list_gadgets(&self) -> Vec<&Gadget> {
        self.0.values().collect()
    }

    pub fn save_gadget(&mut self, gadget: Gadget) -> Result<(), Box<dyn std::error::Error>> {
        let gadget_file_path = file_path()?;
        self.0.insert(gadget.name.clone(), gadget);
        let serialized = serde_json::to_string_pretty(&self.0)?;
        std::fs::write(&gadget_file_path, serialized)?;
        Ok(())
    }

    pub fn delete_gadget(&mut self, name: &str) -> Result<(), Box<dyn std::error::Error>> {
        let gadget_file_path = file_path()?;
        self.0.remove(name);
        let serialized = serde_json::to_string_pretty(&self.0)?;
        std::fs::write(&gadget_file_path, serialized)?;
        Ok(())
    }
}
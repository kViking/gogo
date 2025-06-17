use iced::{Application, Command, Element, Settings, executor, widget::{column, row, button, text, text_input, scrollable, Column, Row, Button, Text, TextInput, Scrollable}};
use gogo_core::{GadgetStore, Gadget};

pub fn main() -> iced::Result {
    GuiApp::run(Settings::default())
}

#[derive(Debug, Clone)]
enum Message {
    SelectGadget(usize),
    ShowAdd,
    AddNameChanged(String),
    AddGadget,
    ShowEdit(usize),
    EditNameChanged(String),
    EditDescChanged(String),
    EditCmdChanged(String),
    SaveEdit,
    DeleteGadget(usize),
    BackToList,
}

struct GuiApp {
    store: GadgetStore,
    gadgets: Vec<Gadget>,
    selected: Option<usize>,
    show_add: bool,
    add_name: String,
    show_edit: bool,
    edit_idx: Option<usize>,
    edit_name: String,
    edit_desc: String,
    edit_cmd: String,
}

impl Application for GuiApp {
    type Executor = executor::Default;
    type Message = Message;
    type Theme = iced::Theme;
    type Flags = ();

    fn new(_flags: ()) -> (Self, Command<Message>) {
        let store = GadgetStore::new().unwrap_or(GadgetStore(Default::default()));
        let gadgets = store.list_gadgets().into_iter().cloned().collect();
        (
            GuiApp {
                store,
                gadgets,
                selected: None,
                show_add: false,
                add_name: String::new(),
                show_edit: false,
                edit_idx: None,
                edit_name: String::new(),
                edit_desc: String::new(),
                edit_cmd: String::new(),
            },
            Command::none(),
        )
    }

    fn title(&self) -> String {
        "GoGoGadget GUI".to_string()
    }

    fn update(&mut self, message: Message) -> Command<Message> {
        match message {
            Message::SelectGadget(idx) => {
                self.selected = Some(idx);
                self.show_add = false;
                self.show_edit = false;
            }
            Message::ShowAdd => {
                self.show_add = true;
                self.add_name.clear();
                self.selected = None;
                self.show_edit = false;
            }
            Message::AddNameChanged(val) => self.add_name = val,
            Message::AddGadget => {
                if !self.add_name.is_empty() {
                    let gadget = Gadget {
                        name: self.add_name.clone(),
                        command: String::new(),
                        description: String::new(),
                        variables: vec![],
                    };
                    let _ = self.store.save_gadget(&gadget);
                    self.gadgets = self.store.list_gadgets().into_iter().cloned().collect();
                    self.show_add = false;
                }
            }
            Message::ShowEdit(idx) => {
                self.show_edit = true;
                self.edit_idx = Some(idx);
                if let Some(g) = self.gadgets.get(idx) {
                    self.edit_name = g.name.clone();
                    self.edit_desc = g.description.clone();
                    self.edit_cmd = g.command.clone();
                }
            }
            Message::EditNameChanged(val) => self.edit_name = val,
            Message::EditDescChanged(val) => self.edit_desc = val,
            Message::EditCmdChanged(val) => self.edit_cmd = val,
            Message::SaveEdit => {
                if let Some(idx) = self.edit_idx {
                    if let Some(g) = self.gadgets.get(idx) {
                        let mut gadget = g.clone();
                        gadget.name = self.edit_name.clone();
                        gadget.description = self.edit_desc.clone();
                        gadget.command = self.edit_cmd.clone();
                        let _ = self.store.save_gadget(&gadget);
                        self.gadgets = self.store.list_gadgets().into_iter().cloned().collect();
                        self.show_edit = false;
                        self.selected = None;
                    }
                }
            }
            Message::DeleteGadget(idx) => {
                if let Some(g) = self.gadgets.get(idx) {
                    let _ = self.store.delete_gadget(&g.name);
                    self.gadgets = self.store.list_gadgets().into_iter().cloned().collect();
                    self.selected = None;
                }
            }
            Message::BackToList => {
                self.selected = None;
                self.show_add = false;
                self.show_edit = false;
            }
        }
        Command::none()
    }

    fn view(&self) -> Element<Message> {
        let mut left_col = column![text("Gadgets").size(24)];
        let mut gadget_list = column![];
        for (i, g) in self.gadgets.iter().enumerate() {
            let row = row![
                button(g.name.as_str()).on_press(Message::SelectGadget(i)),
                button("🗑").on_press(Message::DeleteGadget(i)).style(iced::theme::Button::Destructive)
            ];
            gadget_list = gadget_list.push(row);
        }
        left_col = left_col.push(scrollable(gadget_list).height(iced::Length::Fill)).push(
            button("+ Add Gadget").on_press(Message::ShowAdd)
        );

        let mut content = row![left_col];

        if self.show_add {
            let add_col = column![
                text("Add Gadget").size(20),
                text_input("Name", &self.add_name).on_input(Message::AddNameChanged),
                button("Save").on_press(Message::AddGadget),
                button("Back").on_press(Message::BackToList)
            ].padding(10);
            content = content.push(add_col);
        } else if let Some(idx) = self.selected {
            if let Some(g) = self.gadgets.get(idx) {
                let details_col = column![
                    text(format!("Name: {}", g.name)),
                    text(format!("Description: {}", g.description)),
                    text(format!("Command: {}", g.command)),
                    button("Edit").on_press(Message::ShowEdit(idx)),
                    button("Back").on_press(Message::BackToList)
                ].padding(10);
                content = content.push(details_col);
            }
        }
        if self.show_edit {
            let edit_col = column![
                text("Edit Gadget").size(20),
                text_input("Name", &self.edit_name).on_input(Message::EditNameChanged),
                text_input("Description", &self.edit_desc).on_input(Message::EditDescChanged),
                text_input("Command", &self.edit_cmd).on_input(Message::EditCmdChanged),
                button("Save").on_press(Message::SaveEdit),
                button("Back").on_press(Message::BackToList)
            ].padding(10);
            content = content.push(edit_col);
        }
        content.into()
    }
}

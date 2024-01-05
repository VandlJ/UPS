import tkinter as tk
from GUI.login_window import LoginWindow

if __name__ == "__main__":
    root = tk.Tk()
    client = LoginWindow(root)
    root.mainloop()

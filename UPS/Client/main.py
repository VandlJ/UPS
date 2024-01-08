import tkinter as tk
from GUI.login import LoginScreen

if __name__ == "__main__":
    root = tk.Tk()
    client = LoginScreen(root)
    root.mainloop()

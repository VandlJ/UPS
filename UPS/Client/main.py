import tkinter as tk
from GUI.login import LoginScreen  # Importing the LoginScreen class from the GUI module

if __name__ == "__main__":
    # Main execution block for the application
    root = tk.Tk()  # Creates the main window or root window for the application
    client = LoginScreen(root)  # Initializes the LoginScreen instance using the root window
    root.mainloop()  # Enters the Tkinter event loop to start the GUI application

/*
 * Click nbfs://nbhost/SystemFileSystem/Templates/Licenses/license-default.txt to change this license
 * Click nbfs://nbhost/SystemFileSystem/Templates/Classes/Class.java to edit this template
 */
package themes;

import java.awt.*;
import java.net.URL;
import javax.swing.*;

/**
 *  Clase que sirve de librería gráfica
 * @author Carla Terol 
 */
public class GraphicsHelper {
    
    private class JPanelCustom extends JPanel{
        JPanel p;
        ImageIcon i;
        public JPanelCustom(){
            this.p = new JPanel();
        }
        @Override
        protected void paintComponent(Graphics g) {
            super.paintComponent(g);
            if (this.i != null){
                Image img = this.i.getImage();
                g.drawImage(img, 0, 0, getWidth(), getHeight(), this.p);
            }
            
}
    public URL toURL(String path) {
    URL u;
    u = getClass().getResource(path);
    return u;
}
        /**
         * Para poner la imagen
         * @param filepath ruta al archivo
         */
        public void setI(String filepath){
           
            ImageIcon icon;
            icon = new ImageIcon(
                    getClass().getResource(filepath));
            this.i= icon;}
        }
       
        
}
    
    
    


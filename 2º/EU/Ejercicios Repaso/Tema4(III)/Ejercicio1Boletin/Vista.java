import javax.swing.*;
import java.awt.*;

public class Vista extends JFrame{
    private PanelSuperior panelSuperior;
    private PanelCentral panelCentral;

    public Vista() {
        setTitle("Ejercicio 1 Boletín");
        setSize(500, 300);

        setLayout(new BorderLayout());

        panelSuperior = new PanelSuperior();
        add(panelSuperior, BorderLayout.NORTH);

        panelCentral = new PanelCentral();
        add(panelCentral, BorderLayout.CENTER);
    }

    public class PanelSuperior extends JPanel {
        private JLabel etiqueta;
        private JTextField area;
        private JButton boton;

        public PanelSuperior() {
            etiqueta = new JLabel("Nombre");
            add(etiqueta);

            area = new JTextField(20);
            add(area);

            boton = new JButton("Añadir");
            add(boton);
        }

        public JButton getBoton(){
            return boton;
        }

        public JTextField getField(){
            return area;
        }

        public void setField(String txt){
            area.setText(txt);
        }
    }
    public JButton getBoton(){
        return panelSuperior.getBoton();
    }

    public JTextField getField(){
        return panelSuperior.getField();
    }

    public void setField(String txt){
        panelSuperior.setField(txt);
    }

    public class PanelCentral extends JPanel {
        private JTextArea area1;
        private JScrollPane scroll;

        public PanelCentral(){
            setLayout(new BorderLayout())
            
            area1 = new JTextArea();
            scroll = new JScrollPane(area1);
            add(scroll, BorderLayout.CENTER);
        }

        public JTextArea getTextArea(){
            return area1;
        }

    }

    public JTextArea getTextArea(){
        return panelCentral.getTextArea();
    }

    public void appendTextArea(String txt){
        panelCentral.getTextArea().append(txt + "\n");
    }
}
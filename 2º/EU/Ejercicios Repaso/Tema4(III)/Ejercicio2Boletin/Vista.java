import javax.swing.*;
import java.awt.*;

public class Vista extends JFrame {
    private PanelSuperior panelSuperior;
    private PanelCentral panelCentral;
    private PanelInferior panelInferior;

    public Vista() {
        setTitle("Ejercicio 2 Boletín");
        setSize(500, 300);

        setLayout(new BorderLayout());
        
        panelSuperior = new PanelSuperior();
        add(panelSuperior, BorderLayout.NORTH);
        
        panelCentral = new PanelCentral();
        add(panelCentral, BorderLayout.CENTER);

        panelInferior = new PanelInferior();
        add(panelInferior, BorderLayout.SOUTH);
    }

    public class PanelSuperior extends JPanel {
        private JCheckBox op1, op2, op3;
        private JComboBox<String> items;
        private JLabel lista;
        private String[]  a = {"Item1", "Item2", "Item3"};

        public PanelSuperior() {
            setLayout(new FlowLayout());

            ButtonGroup grupo = new ButtonGroup();
            op1 = new JCheckBox("opcion1");
            grupo.add(op1);
            add(op1);

            op2 = new JCheckBox("opcion2");
            grupo.add(op2);
            add(op2);

            op3 = new JCheckBox("opcion3");
            grupo.add(op3);
            add(op3);

            lista = new JLabel("Lista");
            add(lista);

            items = new JComboBox<>(a);
            add(items);
        }

        public JCheckBox getBoton1(){
            return op1;
        }

        public JCheckBox getBoton2(){
            return op2;
        }

        public JCheckBox getBoton3(){
            return op3;
        } 

        public JComboBox getItems(){
            return items;
        }
    }
    public JCheckBox getBoton1(){
        return panelSuperior.getBoton1();
    }

    public JCheckBox getBoton2(){
        return panelSuperior.getBoton2();
    }

    public JCheckBox getBoton3(){
        return panelSuperior.getBoton3();
    }

    public JComboBox getItems(){
        return panelSuperior.getItems();
    }

    public class PanelCentral extends JPanel {
        private JLabel etiqueta;

        public PanelCentral(){
            etiqueta = new JLabel("10");
            add(etiqueta);
        }

        public JLabel getEtiqueta(){
            return etiqueta;
        }

        public void setEtiqueta(String txt){
            etiqueta.setText(txt);
        }
    }

    public JLabel getEtiqueta(){
        return panelCentral.getEtiqueta();
    }

    public void setEtiqueta(String txt){
        panelCentral.setEtiqueta(txt);
    }

    public class PanelInferior extends JPanel{
        private JButton[] botones;

        public PanelInferior(){
            setLayout(new GridLayout(2,8));

            for(int i = 0; i <= 15; i++){
                botones[i] = new JButton(String.valueOf(i));
                add(botones[i]);
            }
        }
        public JButton[] getBotonesPanel(){
            return botones;
        }
    }
    public JButton[] getBotonesPanel(){
        return panelInferior.getBotonPanel();
    }  
}
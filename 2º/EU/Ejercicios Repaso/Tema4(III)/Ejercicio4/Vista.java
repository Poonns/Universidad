import javax.swing.*;
import java.awt.*;

public class Vista extends JFrame{
    private MiPanel panel;

    public Vista(){
        setTitle("Ejercicio 4");
        setSize(500,300);

        panel = new MiPanel();
        add(panel);
    }

    public class MiPanel extends JPanel{
        private JList<String> lista;
        private String[] ciudades = {"Madrid", "Valencia", "Barcelona", "Sevilla", "Oviedo"};
        private JTextField area;
        private JButton boton;
        private JScrollPane scroll;

        public MiPanel(){
            lista = new JList<>(ciudades);
            
            scroll = new JScrollPane(lista);
            add(scroll);

            area = new JTextField(20);
            add(area);

            boton = new JButton("Un Boton");
            add(boton);

        }

        public JButton getBoton(){
            return boton;
        }

        public JList getLista(){
            return lista;
        }

        public JTextField getField(){
            return area;
        }
    }
    public JButton getBoton(){
        return panel.getBoton();
    }
    public JList getLista(){
        return panel.getLista();
    }
    public JTextField getField(){
        return panel.getField();
    }

}

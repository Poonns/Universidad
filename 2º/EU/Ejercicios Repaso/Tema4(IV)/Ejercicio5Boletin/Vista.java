public class Vista extends JFrame{
    private PanelSuperior panelSuperior;
    private PanelCentral panelCentral;
    private String[] colores = {"Rojo", "Verde", "Azul"};

    public Vista(){
        setTitle("Ejercicio 5 Boletin");
        setSize(600, 400);

        setLayout(new BorderLayout());

        panelSuperior = new PanelSuperior();
        add(panelSuperior, BorderLayout.NORTH);

        panelCentral = new PanelCentral();
        add(panelCentral, BorderLayout.CENTER);
    }

    public class PanelSuperior extends JPanel{
        private JLabel c1, c2, r;
        private JComboBox <String> j1, j2;
        private JSlider slider;

        public PanelSuperior(){
            setBackground(Color.LIGHT_GREY);

            c1 = new JLabel("Color 1");
            add(c1);

            j1 = new JComboBox<>(colores);
            add(j1);

            c2 = new JLabel("Color 2");
            add(c2);

            j2 = new JComboBox<>(colores);
            add(j2);

            r = new JLabel("Radio");
            add(r);

            slider = new JSlider(10,100,50);
            add(slider);

        }

        public String getComBox1(){
            return (String) j1.getSelectedItem();
        }

        public String getComBox2(){
            return (String) j2.getSelectedItem();
        }

        public int getValor(){
            return slider.getValue();
        }
    }

    public String getComBox1(){
        return panelSuperior.getComBox1();
    }

    public String getComBox2(){
        return panelSuperior.getComBox2();
    }

    public int getValor(){
        return panelSuperior.getValor();
    }

    public class PanelCentral extends JPanel{
        Color color1 = Color.RED;
        Color color2 = Color.RED;
        private int radio = 50;

        public void actualizarTamaño(){
            radio = getValor();
            repaint();
        }

        public void actualizarColor(){
            
            if(panelSuperior.getComBox1().equals("Rojo")){
                color1 = Color.RED;
            } else if(panelSuperior.getComBox1().equals("Verde")){
                color1 = Color.GREEN;
            } else if(panelSuperior.getComBox1().equals("Azul")){
                color1 = Color.BLUE;
            } 
            
            if(panelSuperior.getComBox2().equals("Rojo")){
                color2 = Color.RED;
            } else if(panelSuperior.getComBox2().equals("Verde")){
                color2 = Color.GREEN;
            } else if(panelSuperior.getComBox2().equals("Azul")){
                color2 = Color.BLUE;
            }
            repaint();
        }

        public void paintComponent(Graphics g){
            super.paintComponent(g);

            Graphics2D g2d = (Graphics2D) g;

            g2d.setPaint(new GradientPaint(0,0, color1,  getWidth(), 0, color2));
            g2d.fillOval((getWidth() - (radio * 2))/2, (getHeight() - (radio * 2)) /2, radio * 2, radio * 2);
        }
    }
}
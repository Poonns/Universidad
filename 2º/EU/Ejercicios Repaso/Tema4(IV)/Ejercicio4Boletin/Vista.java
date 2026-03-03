public class Vista extends JFrame{
    private MiPanel panel;

    public Vista(){
        setTitle("Ejercicio 4 Boletin");
        setSize(600, 600);

        setLayout(new BorderLayout());

        panel = new MiPanel();
        add(panel, BorderLayout.CENTER);
    }

    public class MiPanel extends JPanel{
        private Color colorActual = Color.RED;

        public void paintComponent(Graphics g){
            super.paintComponent(g);

            g.setColor(colorActual);
            g.fillOval(getWidth() - 100 / 2, getHeight() - 100 / 2, 100, 100);
        }

        public void setColor(Color color){
            colorActual = color;
            repaint();
        }
        
    }
    public void setColor(Color color){
        panel.setColor(color);
    }

    public MiPanel getPanel(){
        return panel;
    }
}
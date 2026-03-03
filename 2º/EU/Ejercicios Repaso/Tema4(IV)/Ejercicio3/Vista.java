public class Vista extends JFrame{
    private MiPanel panel;

    public Vista(){
        setTitle("Ejercicio 3");
        setSize(600, 600);

        setLayout(new BorderLayout());

        panel = new MiPanel();
        add(panel, BorderLayout.CENTER);
    }

    public class MiPanel extends JPanel{
        float rotacion = 0;
        float translacion = 0;

        public void paintComponent(Graphics g){
            super.paintComponent(g);
            Graphics2D g2d = (Graphics2D) g;

            g2d.translate(translacion, translacion);
            g2d.rotate(Math.toRadians(rotacion));

            g2d.setPaint(new GradientPaint(100, 100, Color.GREEN, 200, 400, Color.DARK_GREEN));
            g2d.fillRect(100, 100, 100, 300);
        }

        public void setTecla(char letra){
            if(letra == 'r' || letra == 'R'){
                rotacion += 10;
                repaint();
            } else if(letra == 't' || letra == 'T'){
                translacion += 10;
                repaint();
            }
        }
    }
    public MiPanel getPanel(){
        return panel;
    }

    public void setTecla(char letra){
        panel.setTecla(letra);
    }
}
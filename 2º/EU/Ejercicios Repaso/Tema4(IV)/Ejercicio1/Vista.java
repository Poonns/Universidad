public class Vista extends JFrame {
    private MiPanel panel;
    public Vista(){
        setTitle("Ejercicio 1");
        setSize(600, 600);

        setLayout(new BorderLayout());

        panel = new MiPanel();
        add(panel, BorderLayout.CENTER);
    }
    public class MiPanel extends JPanel{
        public void paintComponent(Graphics g){
            super.paintComponent(g);

            g.setColor(Color.RED);
            g.drawOval(100, 100, 50, 50);

            g.setColor(Color.BLUE);
            g.fillOval((getWidth()/2) - (50/2), (getHeight())/2 - (50/2), 50, 50);
        }
    }
}
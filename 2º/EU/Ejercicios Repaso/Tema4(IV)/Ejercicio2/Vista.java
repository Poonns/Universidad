public class Vista extends JFrame {
    private MiPanel panel;
    public Vista(){
        setTitle("Ejercicio 2");
        setSize(600, 600);

        setLayout(new BorderLayout());

        panel = new MiPanel();
        add(panel, BorderLayout.CENTER);
    }

    public class MiPanel extends JPanel{
        private int PosX = 50;
        private int PosY = 50;

        
        public void paintComponent(Graphics g){
            super.paintComponent(g);

            g.setFont(new Font("Arial", Font.BOLD, 20));

            g.drawString("Posicion: " + PosX + ", " + PosY, PosX, PosY);   
        }

        public void setPosRaton(int x, int y){
            PosX = x;
            PosY = y;
            repaint();
        }
    }
    public void setPosRaton(int x, int y){
        panel.setPosRaton(x, y);
    }

    public MiPanel getPanel(){
        return panel;
    }
}